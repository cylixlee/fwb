package infra

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/cavaliergopher/grab/v3"
	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

type downloader struct {
	http *http.Client
	grab *grab.Client
}

// NewDownloader returns the HTTP downloader used for SDK archives.
func NewDownloader() domain.Downloader {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: consts.HTTPResponseHeaderTimeout,
		},
	}
	grabClient := grab.NewClient()
	grabClient.HTTPClient = client
	grabClient.UserAgent = consts.DownloadUserAgent
	return &downloader{http: client, grab: grabClient}
}

func (d *downloader) Bytes(ctx context.Context, url string) ([]byte, error) {
	var body bytes.Buffer
	err := requests.URL(url).
		Client(d.http).
		ToBytesBuffer(&body).
		Fetch(ctx)
	if err != nil {
		if responseErr, ok := errors.AsType[*requests.ResponseError](err); ok {
			return nil, domain.NewError(consts.ErrorCodeHTTPStatus, url, responseErr.Status)
		}
		return nil, err
	}
	return body.Bytes(), nil
}

func (d *downloader) File(ctx context.Context, url, dest string, progress func(done, total int64)) error {
	if err := os.MkdirAll(filepath.Dir(dest), consts.DirectoryPermissions); err != nil {
		return err
	}
	part := dest + consts.PartFileSuffix
	req, err := grab.NewRequest(part, url)
	if err != nil {
		return err
	}
	response := d.grab.Do(req.WithContext(ctx))

	finished := false
	defer func() {
		if !finished {
			_ = os.Remove(part)
		}
	}()

	ticker := time.NewTicker(consts.DownloadProgressInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if progress != nil {
				progress(response.BytesComplete(), response.Size())
			}
		case <-response.Done:
			if err := response.Err(); err != nil {
				return err
			}
			if err := os.Rename(part, dest); err != nil {
				return err
			}
			finished = true
			return nil
		case <-ctx.Done():
			response.Cancel()
			return ctx.Err()
		}
	}
}
