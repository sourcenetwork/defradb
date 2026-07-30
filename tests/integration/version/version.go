// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

/*
Package version resolves a DefraDB version string (e.g. "v1.0.0") to a path
to an executable release binary for the current platform, downloading and
caching it on demand.

This exists so P2P cross-version compatibility tests can drive a pre-built
release binary as a separate process alongside a native HEAD node.
*/
package version

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	// repo is the GitHub repository that release assets are downloaded from.
	repo = "sourcenetwork/defradb"

	// cacheDirName is the subdirectory of the user cache dir that downloaded
	// binaries are stored under.
	cacheDirName = "defradb-crossversion"
)

// downloader fetches the named release asset for the given tag into destDir.
// It is a seam for testing; the default implementation is an HTTP download.
type downloader func(ctx context.Context, tag string, asset string, destDir string) error

// Provider resolves DefraDB version strings to executable binary paths,
// downloading and caching release assets as needed. The zero value is not
// usable; construct one with [NewProvider].
type Provider struct {
	// cacheDir is the root directory that binaries are cached under.
	cacheDir string

	// goos and goarch are the target platform, defaulting to the running
	// platform's runtime.GOOS/runtime.GOARCH. Overridable for testing.
	goos   string
	goarch string

	// download fetches a release asset. Defaults to httpDownload.
	download downloader
}

// NewProvider returns a Provider that caches binaries under the given
// directory, using the current platform and downloading over HTTP.
func NewProvider(cacheDir string) *Provider {
	return &Provider{
		cacheDir: cacheDir,
		goos:     runtime.GOOS,
		goarch:   runtime.GOARCH,
		download: httpDownload,
	}
}

// defaultProvider is the Provider used by the package-level [BinaryPath].
func defaultProvider() (*Provider, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.Wrap("failed to resolve user cache dir", err)
	}
	return NewProvider(filepath.Join(userCacheDir, cacheDirName)), nil
}

// BinaryPath returns a filesystem path to an executable defradb binary of the
// given version for the current platform, downloading and caching it if
// necessary. skip is true when no release asset exists for this GOOS/GOARCH
// (the caller should skip, not fail). err is only returned for genuine
// failures, e.g. a download error.
func BinaryPath(ctx context.Context, version string) (path string, skip bool, err error) {
	p, err := defaultProvider()
	if err != nil {
		return "", false, err
	}
	return p.BinaryPath(ctx, version)
}

// BinaryPath returns a filesystem path to an executable defradb binary of the
// given version for the Provider's platform, downloading and caching it if
// necessary. skip is true when no release asset exists for this GOOS/GOARCH
// (the caller should skip, not fail). err is only returned for genuine
// failures, e.g. a download error.
func (p *Provider) BinaryPath(ctx context.Context, version string) (path string, skip bool, err error) {
	asset, ok := assetName(version, p.goos, p.goarch)
	if !ok {
		return "", true, nil
	}

	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	destDir := filepath.Join(p.cacheDir, tag)
	destPath := filepath.Join(destDir, asset)

	if isExecutable(destPath) {
		return destPath, false, nil
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", false, errors.Wrap("failed to create cache dir", err, errors.NewKV("dir", destDir))
	}

	if err := p.download(ctx, tag, asset, destDir); err != nil {
		return "", false, errors.Wrap(
			"failed to download release asset",
			err,
			errors.NewKV("tag", tag),
			errors.NewKV("asset", asset),
		)
	}

	if err := os.Chmod(destPath, 0o755); err != nil {
		return "", false, errors.Wrap("failed to make binary executable", err, errors.NewKV("path", destPath))
	}

	return destPath, false, nil
}

// assetName returns the name of the release asset for the given version and
// platform, and whether that platform has a published asset.
func assetName(version, goos, goarch string) (name string, ok bool) {
	var osName string
	switch goos {
	case "linux", "darwin", "windows":
		osName = goos
	default:
		return "", false
	}

	var archName string
	switch goarch {
	case "amd64":
		archName = "x86_64"
	case "arm64":
		archName = "arm64"
	default:
		return "", false
	}

	ver := strings.TrimPrefix(version, "v")
	name = fmt.Sprintf("defradb_%s_%s_%s", ver, osName, archName)
	if goos == "windows" {
		name += ".exe"
	}
	return name, true
}

// isExecutable reports whether path exists and has at least one executable
// bit set.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}

// httpDownload downloads the named release asset for tag into destDir. Release
// assets of a public repo are served without authentication, so no token is
// needed. The asset is streamed to a temp file and renamed on success, so a
// failed download never leaves a partial file at the final path.
func httpDownload(ctx context.Context, tag string, asset string, destDir string) error {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status downloading release asset",
			errors.NewKV("url", url), errors.NewKV("status", res.Status))
	}

	tmp, err := os.CreateTemp(destDir, asset+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, res.Body); err != nil {
		tmp.Close()        //nolint:errcheck
		os.Remove(tmpPath) //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath) //nolint:errcheck
		return err
	}

	if err := os.Rename(tmpPath, filepath.Join(destDir, asset)); err != nil {
		os.Remove(tmpPath) //nolint:errcheck
		return err
	}
	return nil
}
