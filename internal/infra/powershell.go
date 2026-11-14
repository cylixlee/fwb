// Package infra implements the ports declared by the domain layer. All
// Windows interaction goes through WSL interop + embedded PowerShell templates;
// fwb has no Windows-side helper.
package infra

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode/utf16"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	flutterpkg "github.com/cylixlee/fwb/internal/infra/flutter"
	"github.com/cylixlee/fwb/pkg/fsx"
)

//go:embed scripts
var scripts embed.FS

//go:embed scripts/preamble.ps1
var preambleScript string

var (
	psOnce sync.Once
	psPath string
	psErr  error

	tmplOnce sync.Once
	tmpl     *template.Template
	tmplErr  error
)

// quote returns a single-quoted PowerShell string literal.
func quote(s string) string {
	return consts.SingleQuote + strings.ReplaceAll(s, consts.SingleQuote, consts.EscapedSingleQuote) + consts.SingleQuote
}

func lookPowerShell() (string, error) {
	psOnce.Do(func() {
		var candidates []string
		if p, err := exec.LookPath(consts.PowerShellExecutable); err == nil {
			candidates = append(candidates, p)
		}
		if p, err := exec.LookPath(consts.PwshExecutable); err == nil {
			candidates = append(candidates, p)
		}
		candidates = append(candidates, consts.FallbackPowerShellPath, consts.FallbackPwshPath)
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				psPath = c
				return
			}
		}
		psErr = domain.NewError(consts.ErrorCodeWinNoPowerShell)
	})
	return psPath, psErr
}

func renderScript(name string, data any) (string, error) {
	raw, err := scripts.ReadFile(consts.ScriptsDirectory + consts.EmbedPathSeparator + name)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(name, consts.ScriptTemplateSuffix) {
		return string(raw), nil
	}
	tmplOnce.Do(func() {
		tmpl, tmplErr = template.New(consts.TemplateRootName).
			Funcs(template.FuncMap{consts.TemplateFunctionQuote: quote}).
			ParseFS(scripts, consts.ScriptsGlob)
	})
	if tmplErr != nil {
		return "", tmplErr
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinScript, err, name)
	}
	return buf.String(), nil
}

func encodeCommand(script string) string {
	u := utf16.Encode([]rune(preambleScript + script))
	buf := make([]byte, 0, len(u)*2)
	for _, r := range u {
		buf = append(buf, byte(r), byte(r>>8))
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func psCommand(ctx context.Context, script string) (*exec.Cmd, error) {
	ps, err := lookPowerShell()
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, ps,
		consts.PowerShellFlagNoProfile,
		consts.PowerShellFlagNonInteractive,
		consts.PowerShellFlagEncodedCommand,
		encodeCommand(script),
	), nil
}

func runPowerShell(ctx context.Context, script string) (string, error) {
	cmd, err := psCommand(ctx, script)
	if err != nil {
		return "", err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			err = fmt.Errorf(consts.ErrorCauseFormat, msg, err)
		}
		return stdout.String(), domain.WrapError(consts.ErrorCodeWinPowerShell, err)
	}
	return stdout.String(), nil
}

func runPowerShellName(ctx context.Context, name string, data any) (string, error) {
	script, err := renderScript(name, data)
	if err != nil {
		return "", err
	}
	return runPowerShell(ctx, script)
}

func runPowerShellStream(ctx context.Context, name string, data any, stdout, stderr io.Writer) error {
	script, err := renderScript(name, data)
	if err != nil {
		return err
	}
	cmd, err := psCommand(ctx, script)
	if err != nil {
		return err
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return domain.WrapError(consts.ErrorCodeWinPowerShell, err)
	}
	return nil
}

func toWindowsPath(ctx context.Context, p string) (string, error) {
	out, err := exec.CommandContext(ctx, consts.WslPathExecutable, consts.WslPathWindowsFlag, p).Output()
	if err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinPath, err, consts.WslPathWindowsFlag, p)
	}
	return strings.TrimSpace(string(out)), nil
}

func toWSLPath(ctx context.Context, p string) (string, error) {
	out, err := exec.CommandContext(ctx, consts.WslPathExecutable, consts.WslPathWSLFlag, p).Output()
	if err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinPath, err, consts.WslPathWSLFlag, p)
	}
	return strings.TrimSpace(string(out)), nil
}

// powershellHost implements domain.WindowsHost.
type powershellHost struct{}

// NewWindowsHost returns the Windows interop gateway.
func NewWindowsHost() domain.WindowsHost { return powershellHost{} }

func (powershellHost) Available() error {
	if _, err := lookPowerShell(); err != nil {
		return err
	}
	if _, err := exec.LookPath(consts.WslPathExecutable); err != nil {
		return domain.NewError(consts.ErrorCodeWinNoWslpath)
	}
	return nil
}

func (powershellHost) Info(ctx context.Context) (string, error) {
	out, err := runPowerShellName(ctx, consts.ScriptPowerShellInfo, nil)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(strings.TrimSpace(out), consts.CarriageReturnLineFeed, consts.InfoSeparator), nil
}

type envParams struct {
	LocalAppData string
	UserProfile  string
}

func (powershellHost) Env(ctx context.Context) (domain.WindowsEnvironment, error) {
	out, err := runPowerShellName(ctx, consts.ScriptEnvironment, envParams{
		LocalAppData: consts.EnvLocalAppData,
		UserProfile:  consts.EnvUserProfile,
	})
	if err != nil {
		return domain.WindowsEnvironment{}, err
	}
	lines := strings.Split(strings.ReplaceAll(out, consts.CarriageReturnLineFeed, consts.LineFeed), consts.LineFeed)
	if len(lines) < 2 {
		return domain.WindowsEnvironment{}, domain.NewError(consts.ErrorCodeWinEnv, out)
	}
	env := domain.WindowsEnvironment{
		LocalAppDataWin: strings.TrimSpace(lines[0]),
		OSVersion:       strings.TrimSpace(strings.Join(lines[2:], consts.InfoSeparator)),
	}
	userProfile := strings.TrimSpace(lines[1])
	if env.LocalAppDataWin == "" || userProfile == "" {
		return domain.WindowsEnvironment{}, domain.NewError(consts.ErrorCodeWinEnv, out)
	}
	if env.LocalAppDataWSL, err = toWSLPath(ctx, env.LocalAppDataWin); err != nil {
		return domain.WindowsEnvironment{}, err
	}
	if env.UserProfileWSL, err = toWSLPath(ctx, userProfile); err != nil {
		return domain.WindowsEnvironment{}, err
	}
	return env, nil
}

func (powershellHost) Mirrored(ctx context.Context) (bool, string, error) {
	env, err := cachedEnv(ctx)
	if err != nil {
		return false, "", err
	}
	path := filepath.Join(env.UserProfileWSL, consts.WSLConfigFileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, path, nil
	}
	if err != nil {
		return false, path, err
	}
	for line := range strings.SplitSeq(strings.ToLower(string(data)), consts.LineFeed) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, consts.WSLNetworkingModeKey) && strings.Contains(line, consts.WSLMirrorModeValue) {
			return true, path, nil
		}
	}
	return false, path, nil
}

func (powershellHost) ToWindows(ctx context.Context, path string) (string, error) {
	return toWindowsPath(ctx, path)
}

func (powershellHost) ToWSL(ctx context.Context, path string) (string, error) {
	return toWSLPath(ctx, path)
}

type versionParams struct {
	Bat         string
	VersionFlag string
	MachineFlag string
}

func (powershellHost) FlutterVersion(ctx context.Context, flutterBatWindows string) (domain.FlutterVersion, error) {
	out, err := runPowerShellName(ctx, consts.ScriptFlutterVersion, versionParams{
		Bat:         flutterBatWindows,
		VersionFlag: consts.FlutterVersionFlag,
		MachineFlag: consts.FlutterMachineFlag,
	})
	if err != nil {
		return domain.FlutterVersion{}, err
	}
	return flutterpkg.Parse(out)
}

type extractParams struct {
	Zip             string
	Dest            string
	Tar             string
	ExpandArchive   string
	TarFailedFormat string
}

func (powershellHost) Extract(ctx context.Context, zipWSL, destWSL string) error {
	zipWin, err := toWindowsPath(ctx, zipWSL)
	if err != nil {
		return err
	}
	destWin, err := toWindowsPath(ctx, destWSL)
	if err != nil {
		return err
	}
	_, err = runPowerShellName(ctx, consts.ScriptExtract, extractParams{
		Zip:             zipWin,
		Dest:            destWin,
		Tar:             consts.TarExecutable,
		ExpandArchive:   consts.ExpandArchive,
		TarFailedFormat: consts.TarFailedFormat,
	})
	return err
}

type buildParams struct {
	PubCache          string
	PubHostedURL      string
	StorageBaseURL    string
	PubCacheEnv       string
	PubHostedURLEnv   string
	StorageBaseURLEnv string
	Mirror            string
	Bat               string
	Args              string
}

func (h powershellHost) Build(ctx context.Context, req domain.WindowsBuildRequest, stdout, stderr io.Writer) error {
	mirror, err := toWindowsPath(ctx, req.Mirror)
	if err != nil {
		return err
	}
	data := buildParams{
		PubCache:          req.Env[consts.EnvPubCache],
		PubHostedURL:      req.Env[consts.EnvPubHostedURL],
		StorageBaseURL:    req.Env[consts.EnvFlutterStorageBaseURL],
		PubCacheEnv:       consts.EnvPubCache,
		PubHostedURLEnv:   consts.EnvPubHostedURL,
		StorageBaseURLEnv: consts.EnvFlutterStorageBaseURL,
		Mirror:            mirror,
		Bat:               req.FlutterBatWindows,
		Args:              consts.FlutterWindowsBuildArgs,
	}
	return runPowerShellStream(ctx, consts.ScriptBuild, data, stdout, stderr)
}

type runParams struct {
	Dir                  string
	Exe                  string
	SessionFile          string
	Switches             []string
	SwitchesVariable     string
	SwitchVariablePrefix string
	SessionEncoding      string
}

func (h powershellHost) Run(ctx context.Context, req domain.WindowsRunRequest, stdout, stderr io.Writer) error {
	dirWin, err := toWindowsPath(ctx, req.Dir)
	if err != nil {
		return err
	}
	exeWin, err := toWindowsPath(ctx, req.Exe)
	if err != nil {
		return err
	}
	sessionWin, err := toWindowsPath(ctx, req.SessionFile)
	if err != nil {
		return err
	}
	data := runParams{
		Dir:                  dirWin,
		Exe:                  exeWin,
		SessionFile:          sessionWin,
		Switches:             req.Switches,
		SwitchesVariable:     consts.EngineSwitchesVariable,
		SwitchVariablePrefix: consts.EngineSwitchVariablePrefix,
		SessionEncoding:      consts.SessionFileEncoding,
	}
	return runPowerShellStream(ctx, consts.ScriptRun, data, stdout, stderr)
}

type killParams struct {
	PID                      int
	CloseTimeoutMilliseconds int
}

func (powershellHost) Stop(ctx context.Context, pid int) error {
	script, err := renderScript(consts.ScriptKill, killParams{PID: pid, CloseTimeoutMilliseconds: consts.ProcessCloseTimeoutMillis})
	if err != nil {
		return err
	}
	_, err = runPowerShell(ctx, script)
	return err
}

// cachedEnv caches the Windows environment for 12h in ~/.cache/fwb/winenv.json.
func cachedEnv(ctx context.Context) (domain.WindowsEnvironment, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return domain.WindowsEnvironment{}, err
	}
	dir := filepath.Join(cacheDir, consts.CacheDirectoryName)
	if err := os.MkdirAll(dir, consts.DirectoryPermissions); err != nil {
		return domain.WindowsEnvironment{}, err
	}
	path := filepath.Join(dir, consts.WinEnvCacheFileName)
	if data, err := os.ReadFile(path); err == nil {
		var cached struct {
			domain.WindowsEnvironment
			FetchedAt int64
		}
		if json.Unmarshal(data, &cached) == nil &&
			cached.LocalAppDataWSL != "" &&
			time.Since(time.Unix(cached.FetchedAt, 0)) < consts.WinEnvCacheTTL {
			if st, err := os.Stat(cached.LocalAppDataWSL); err == nil && st.IsDir() {
				return cached.WindowsEnvironment, nil
			}
		}
	}
	env, err := NewWindowsHost().Env(ctx)
	if err != nil {
		return domain.WindowsEnvironment{}, err
	}
	payload := map[string]any{
		consts.JSONLocalAppDataWSL: env.LocalAppDataWSL,
		consts.JSONLocalAppDataWin: env.LocalAppDataWin,
		consts.JSONUserProfileWSL:  env.UserProfileWSL,
		consts.JSONOSVersion:       env.OSVersion,
		consts.JSONFetchedAt:       time.Now().Unix(),
	}
	_ = fsx.WriteJSONAtomic(path, payload, consts.FilePermissions)
	return env, nil
}
