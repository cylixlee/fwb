package infra

import (
	"bytes"
	"context"
	"crypto/tls"
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
			// Large archives are more reliable over HTTP/1.1; mirrors have
			// been observed resetting long-lived HTTP/2 streams.
			ForceAttemptHTTP2: false,
			TLSNextProto:      make(map[string]func(string, *tls.Conn) http.RoundTripper),
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
	var lastErr error
	for attempt := 1; attempt <= consts.DownloadAttempts; attempt++ {
		lastErr = d.download(ctx, url, dest, progress)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil || attempt == consts.DownloadAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return lastErr
		case <-time.After(consts.DownloadRetryDelay):
		}
	}
	return lastErr
}

// download keeps the .part file between attempts so grab can resume.
func (d *downloader) download(ctx context.Context, url, dest string, progress func(done, total int64)) error {
	part := dest + consts.PartFileSuffix
	req, err := grab.NewRequest(part, url)
	if err != nil {
		return err
	}
	response := d.grab.Do(req.WithContext(ctx))

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
			// Report the final byte count so the last frame is accurate.
			if progress != nil {
				progress(response.BytesComplete(), response.Size())
			}
			return os.Rename(part, dest)
		case <-ctx.Done():
			response.Cancel()
			return ctx.Err()
		}
	}
}
