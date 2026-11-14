package flutter

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

type releaseSource struct {
	dl domain.Downloader
}

// NewReleaseSource resolves official Windows portable SDKs.
func NewReleaseSource(dl domain.Downloader) domain.WindowsSDKReleaseSource {
	return &releaseSource{dl: dl}
}

func (r *releaseSource) Find(ctx context.Context, v domain.FlutterVersion, storageBase string) (domain.WindowsSDKRelease, error) {
	if storageBase == "" {
		storageBase = consts.DefaultStorageBaseURL
	}
	url := strings.TrimRight(storageBase, consts.WSLPathPrefix) + consts.ReleaseIndexPath
	body, err := r.dl.Bytes(ctx, url)
	if err != nil {
		return domain.WindowsSDKRelease{}, domain.WrapError(consts.ErrorCodeFlutterReleaseIndex, err, url)
	}
	var index struct {
		Releases []struct {
			Version string `json:"version"`
			Channel string `json:"channel"`
			Archive string `json:"archive"`
			SHA256  string `json:"sha256"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return domain.WindowsSDKRelease{}, domain.WrapError(consts.ErrorCodeFlutterReleaseParse, err, url)
	}
	if len(index.Releases) == 0 {
		return domain.WindowsSDKRelease{}, domain.NewError(consts.ErrorCodeFlutterReleaseEmpty, url)
	}
	var fallback *domain.WindowsSDKRelease
	for _, r := range index.Releases {
		if r.Version != v.Framework {
			continue
		}
		release := domain.WindowsSDKRelease{
			Version: r.Version,
			Channel: r.Channel,
			Archive: r.Archive,
			SHA256:  r.SHA256,
		}
		if strings.EqualFold(r.Channel, v.Channel) {
			return release, nil
		}
		if fallback == nil {
			fallback = &release
		}
	}
	if fallback != nil {
		return *fallback, nil
	}
	return domain.WindowsSDKRelease{}, domain.NewError(consts.ErrorCodeFlutterReleaseNotFound, v.Framework, v.Channel)
}
