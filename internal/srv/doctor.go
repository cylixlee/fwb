package srv

import (
	"context"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// Diagnostics runs the doctor probes for `fwb doctor`.
type Diagnostics struct {
	cfg     domain.ConfigStore
	host    domain.WindowsHost
	flutter domain.WSLFlutterCLI
	mirrors domain.ProjectMirrorFactory
	sdks    domain.SDKStore
}

// NewDiagnostics wires the doctor use case.
func NewDiagnostics(
	cfg domain.ConfigStore,
	host domain.WindowsHost,
	flutter domain.WSLFlutterCLI,
	mirrors domain.ProjectMirrorFactory,
	sdks domain.SDKStore,
) *Diagnostics {
	return &Diagnostics{cfg: cfg, host: host, flutter: flutter, mirrors: mirrors, sdks: sdks}
}

// Collect runs the doctor table in report order.
func (d *Diagnostics) Collect(ctx context.Context) domain.DoctorReport {
	p := &probes{diagnostics: d, ctx: ctx}
	p.version, p.versionErr = d.flutter.Version(ctx)

	var report domain.DoctorReport
	for _, diagnose := range p.table() {
		entries, stop := diagnose()
		report.Entries = append(report.Entries, entries...)
		if stop {
			break
		}
	}
	return report
}

// probes is the state shared by the checks of one doctor run.
type probes struct {
	diagnostics *Diagnostics
	ctx         context.Context
	version     domain.FlutterVersion
	versionErr  error
}

// probe is one diagnostic check. It returns its entries and whether the
// remaining checks are pointless (no PowerShell, no Windows diagnostics).
type probe func() (entries []domain.DoctorEntry, stop bool)

// table lists the probes in report order.
func (p *probes) table() []probe {
	return []probe{
		p.wsl,
		p.powerShell,
		p.windows,
		p.network,
		p.mirrorRoot,
		p.override,
		p.sdk,
		p.config,
	}
}

func (p *probes) wsl() ([]domain.DoctorEntry, bool) {
	var e entries
	if p.versionErr != nil {
		e.fail(consts.DoctorWSLFailure, p.versionErr)
	} else {
		e.ok(consts.DoctorWSLSuccess, p.version.Framework, p.version.Channel, p.version.ShortEngine())
	}
	return e.list, false
}

func (p *probes) powerShell() ([]domain.DoctorEntry, bool) {
	var e entries
	if err := p.diagnostics.host.Available(); err != nil {
		e.fail(consts.DoctorPowerShellFailure, err)
		return e.list, true
	}
	if info, err := p.diagnostics.host.Info(p.ctx); err != nil {
		e.fail(consts.DoctorPowerShellFailure, err)
	} else {
		e.ok(consts.DoctorPowerShellSuccess, info)
	}
	return e.list, false
}

func (p *probes) windows() ([]domain.DoctorEntry, bool) {
	var e entries
	env, err := p.diagnostics.host.Env(p.ctx)
	if err != nil {
		e.warn(consts.DoctorWindowsFailure, err)
		return e.list, false
	}
	e.ok(consts.DoctorWindowsVersion, env.OSVersion)
	e.ok(consts.DoctorLocalAppData, env.LocalAppDataWin)
	return e.list, false
}

func (p *probes) network() ([]domain.DoctorEntry, bool) {
	var e entries
	mirrored, path, err := p.diagnostics.host.Mirrored(p.ctx)
	switch {
	case err != nil:
		e.warn(consts.DoctorMirroredFailure, err)
	case mirrored:
		e.ok(consts.DoctorMirroredSuccess, path)
	default:
		e.warn(consts.DoctorMirroredWarning, path)
		e.info(consts.DoctorMirroredHint, path)
	}
	return e.list, false
}

func (p *probes) mirrorRoot() ([]domain.DoctorEntry, bool) {
	var e entries
	root, err := p.diagnostics.mirrors.Root(p.ctx)
	if err != nil {
		e.warn(consts.DoctorMirrorRootFailure, err)
	} else {
		e.ok(consts.DoctorMirrorRoot, root)
	}
	return e.list, false
}

func (p *probes) override() ([]domain.DoctorEntry, bool) {
	var e entries
	if override := p.diagnostics.cfg.Get(consts.ConfigKeyWindowsFlutter); override != "" {
		e.info(consts.DoctorOverride, override)
	}
	return e.list, false
}

func (p *probes) sdk() ([]domain.DoctorEntry, bool) {
	var e entries
	if p.versionErr != nil {
		return e.list, false
	}
	if sdk, cached := p.diagnostics.sdks.Cached(p.ctx, p.version); cached {
		e.ok(consts.DoctorSDKCached, sdk.FlutterRoot)
	} else {
		e.info(consts.DoctorSDKMissing)
	}
	return e.list, false
}

func (p *probes) config() ([]domain.DoctorEntry, bool) {
	var e entries
	e.info(consts.DoctorConfigPath, p.diagnostics.cfg.Path())
	return e.list, false
}

// entries collects the lines one probe contributes.
type entries struct {
	list []domain.DoctorEntry
}

func (e *entries) add(level consts.DoctorLevel, code string, args ...any) {
	e.list = append(e.list, domain.DoctorEntry{Level: level, Code: code, Args: args})
}

func (e *entries) ok(code string, args ...any) {
	e.add(consts.DoctorLevelOK, code, args...)
}

func (e *entries) warn(code string, args ...any) {
	e.add(consts.DoctorLevelWarn, code, args...)
}

func (e *entries) info(code string, args ...any) {
	e.add(consts.DoctorLevelInfo, code, args...)
}

func (e *entries) fail(code string, args ...any) {
	e.add(consts.DoctorLevelFail, code, args...)
}
