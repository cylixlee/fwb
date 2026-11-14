package powershell

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
)

// Environment returns the Windows environment, cached for 12h in
// ~/.cache/fwb/winenv.json.
func Environment(ctx context.Context) (domain.WindowsEnvironment, error) {
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
	env, err := NewHost().Env(ctx)
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
	fsx.WriteJSONAtomic(path, payload, consts.FilePermissions)
	return env, nil
}
