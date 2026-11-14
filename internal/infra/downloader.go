package infra

import (
	"context"
	"errors"
	"fmt"
	"github.com/cylixlee/fwb/internal/consts"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cylixlee/fwb/internal/domain"
)

type httpDownloader struct {
	client *http.Client
}

// NewDownloader returns the HTTP downloader used for SDK archives.
func NewDownloader() domain.Downloader {
	return &httpDownloader{client: &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: consts.HTTPResponseHeaderTimeout,
		},
	}}
}

func (d *httpDownloader) Bytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewError(consts.ErrorCodeHTTPStatus, url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (d *httpDownloader) File(ctx context.Context, url, dest string, progress func(done, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.NewError(consts.ErrorCodeHTTPStatus, url, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(dest), consts.DirectoryPermissions); err != nil {
		return err
	}
	part := dest + consts.PartFileSuffix
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(part)
		}
	}()
	writeErr := func() (err error) {
		out, err := os.Create(part)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := out.Close(); err == nil {
				err = cerr
			}
		}()

		total := resp.ContentLength
		var done int64
		buf := make([]byte, consts.DownloadBufferSize)
		for {
			n, rerr := resp.Body.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					return werr
				}
				done += int64(n)
				if progress != nil {
					progress(done, total)
				}
			}
			if errors.Is(rerr, io.EOF) {
				break
			}
			if rerr != nil {
				return rerr
			}
		}
		return out.Sync()
	}()
	if writeErr != nil {
		return writeErr
	}
	if err := os.Rename(part, dest); err != nil {
		return fmt.Errorf(consts.ErrorCauseFormat, dest, err)
	}
	renamed = true
	return nil
}
