package powershell

import (
	"context"
	"io"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

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

func (h host) Build(ctx context.Context, req domain.WindowsBuildRequest, stdout, stderr io.Writer) error {
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
