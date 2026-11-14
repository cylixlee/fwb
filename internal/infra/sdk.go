package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/infra/mirror"
	"github.com/cylixlee/fwb/pkg/fsx"
)

type sdkStore struct {
	cfg  domain.ConfigStore
	host domain.WindowsHost
}

// NewSDKStore returns the Windows SDK cache.
func NewSDKStore(cfg domain.ConfigStore, host domain.WindowsHost) domain.SDKStore {
	return &sdkStore{cfg: cfg, host: host}
}

func (s *sdkStore) Root(ctx context.Context) (string, error) {
	root, err := mirror.Root(ctx, s.cfg, s.host)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, consts.SDKDirectoryName), nil
}

func sdkDirName(v domain.FlutterVersion) string {
	channel := v.Channel
	if channel == "" {
		channel = consts.DefaultChannel
	}
	return v.Framework + consts.SDKDirectorySeparator + channel
}

func (s *sdkStore) Cached(ctx context.Context, v domain.FlutterVersion) (*domain.WindowsSDK, bool) {
	root, err := s.Root(ctx)
	if err != nil {
		return nil, false
	}
	sdk := sdkFor(v, filepath.Join(root, sdkDirName(v)))
	if _, err := os.Stat(sdk.FlutterBat()); err != nil {
		return nil, false
	}
	marker, err := os.ReadFile(filepath.Join(sdk.Root, consts.SDKMarkerFileName))
	if err != nil {
		return nil, false
	}
	var got domain.FlutterVersion
	if json.Unmarshal(marker, &got) != nil {
		return nil, false
	}
	if got.EngineRevision != v.EngineRevision || got.Framework != v.Framework {
		return nil, false
	}
	return sdk, true
}

func (s *sdkStore) Prepare(ctx context.Context, v domain.FlutterVersion) (*domain.WindowsSDK, error) {
	root, err := s.Root(ctx)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, sdkDirName(v))
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, consts.DirectoryPermissions); err != nil {
		return nil, err
	}
	return sdkFor(v, dir), nil
}

func (s *sdkStore) MarkReady(ctx context.Context, v domain.FlutterVersion) error {
	root, err := s.Root(ctx)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, sdkDirName(v))
	if err := os.MkdirAll(dir, consts.DirectoryPermissions); err != nil {
		return err
	}
	return fsx.WriteJSONAtomic(filepath.Join(dir, consts.SDKMarkerFileName), v, consts.FilePermissions)
}

func sdkFor(v domain.FlutterVersion, root string) *domain.WindowsSDK {
	return &domain.WindowsSDK{
		Version:     v,
		Root:        root,
		FlutterRoot: filepath.Join(root, consts.FlutterSDKInnerDirName),
	}
}

func (s *sdkStore) Checksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
