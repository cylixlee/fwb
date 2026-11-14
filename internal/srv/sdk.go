package srv

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// sdkProvider installs and caches Windows Flutter SDKs.
type sdkProvider struct {
	cfg        domain.ConfigStore
	host       domain.WindowsHost
	releases   domain.WindowsSDKReleaseSource
	downloader domain.Downloader
	store      domain.SDKStore
}

// NewSDKProvider wires the SDK provider.
func NewSDKProvider(
	cfg domain.ConfigStore,
	host domain.WindowsHost,
	releases domain.WindowsSDKReleaseSource,
	downloader domain.Downloader,
	store domain.SDKStore,
) domain.SDKProvider {
	return &sdkProvider{cfg: cfg, host: host, releases: releases, downloader: downloader, store: store}
}

// Ensure returns a usable Windows SDK for want, downloading it when needed.
func (p *sdkProvider) Ensure(ctx context.Context, want domain.FlutterVersion, rep domain.EventReporter) (*domain.WindowsSDK, error) {
	if override := strings.TrimSpace(p.cfg.Get(consts.ConfigKeyWindowsFlutter)); override != "" {
		rootWSL := override
		if !strings.HasPrefix(rootWSL, consts.WSLPathPrefix) {
			var err error
			if rootWSL, err = p.host.ToWSL(ctx, override); err != nil {
				return nil, err
			}
		}
		sdk := &domain.WindowsSDK{Version: want, Root: rootWSL, FlutterRoot: rootWSL}
		if _, err := os.Stat(sdk.FlutterBat()); err != nil {
			return nil, domain.NewError(consts.ErrorCodeSDKOverride, override)
		}
		batWin, err := p.host.ToWindows(ctx, sdk.FlutterBat())
		if err != nil {
			return nil, err
		}
		got, err := p.host.FlutterVersion(ctx, batWin)
		if err != nil {
			return nil, domain.WrapError(consts.ErrorCodeSDKVersion, err, override)
		}
		if !got.SameEngine(want) {
			return nil, domain.NewError(consts.ErrorCodeFlutterOverride,
				want.Framework, want.EngineRevision, got.Framework, got.EngineRevision)
		}
		return sdk, nil
	}

	if cached, ok := p.store.Cached(ctx, want); ok {
		return cached, nil
	}

	storage := strings.TrimSpace(p.cfg.Get(consts.ConfigKeyWindowsStorageURL))
	if storage == "" {
		storage = os.Getenv(consts.EnvFlutterStorageBaseURL)
	}
	release, err := p.releases.Find(ctx, want, storage)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(storage, consts.WSLPathPrefix)
	if base == "" {
		base = consts.DefaultStorageBaseURL
	}
	url := base + consts.ReleaseArchivePath + release.Archive

	emit(rep, consts.EventSDKDownloading, want.Framework, want.Channel)
	sdkRoot, err := p.store.Root(ctx)
	if err != nil {
		return nil, err
	}
	archive := filepath.Join(sdkRoot, consts.SDKDownloadsDirectory, filepath.Base(release.Archive))
	// Register the cleanup before the download: a failed or aborted transfer
	// must not leave a multi-gigabyte archive behind.
	defer os.Remove(archive)
	gauge := meter(rep, consts.ProgressSDKDownload, want.Framework)
	if err := p.download(ctx, url, archive, gauge); err != nil {
		return nil, domain.WrapError(consts.ErrorCodeSDKDownload, err, url)
	}
	if release.SHA256 != "" {
		sum, err := p.store.Checksum(archive)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(sum, release.SHA256) {
			return nil, domain.NewError(consts.ErrorCodeSDKChecksum, archive)
		}
	}

	sdk, err := p.store.Prepare(ctx, want)
	if err != nil {
		return nil, err
	}
	emit(rep, consts.EventSDKExtracting, sdk.Root)
	if err := p.host.Extract(ctx, archive, sdk.Root); err != nil {
		return nil, domain.WrapError(consts.ErrorCodeSDKExtract, err)
	}
	if _, err := os.Stat(sdk.FlutterBat()); err != nil {
		return nil, domain.NewError(consts.ErrorCodeSDKMissingBat, sdk.FlutterBat())
	}
	if err := p.store.MarkReady(ctx, want); err != nil {
		return nil, err
	}
	emit(rep, consts.EventSDKReady)
	return sdk, nil
}

// download streams the archive through the meter. The meter is finished on
// every path: a failed transfer aborts it instead of claiming completion.
func (p *sdkProvider) download(ctx context.Context, url, archive string, gauge domain.OperationProgress) (err error) {
	defer func() {
		if err != nil {
			gauge.Abort()
			return
		}
		gauge.Done()
	}()
	return p.downloader.File(ctx, url, archive, gauge.Update)
}
