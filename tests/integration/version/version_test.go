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

package version

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goos    string
		goarch  string
		want    string
		wantOk  bool
	}{
		{
			name:    "linux amd64",
			version: "v1.0.0",
			goos:    "linux",
			goarch:  "amd64",
			want:    "defradb_1.0.0_linux_x86_64",
			wantOk:  true,
		},
		{
			name:    "linux arm64",
			version: "v1.0.0",
			goos:    "linux",
			goarch:  "arm64",
			want:    "defradb_1.0.0_linux_arm64",
			wantOk:  true,
		},
		{
			name:    "darwin arm64",
			version: "v1.0.0",
			goos:    "darwin",
			goarch:  "arm64",
			want:    "defradb_1.0.0_darwin_arm64",
			wantOk:  true,
		},
		{
			name:    "darwin amd64",
			version: "v1.0.0",
			goos:    "darwin",
			goarch:  "amd64",
			want:    "defradb_1.0.0_darwin_x86_64",
			wantOk:  true,
		},
		{
			name:    "windows amd64",
			version: "v1.0.0",
			goos:    "windows",
			goarch:  "amd64",
			want:    "defradb_1.0.0_windows_x86_64.exe",
			wantOk:  true,
		},
		{
			name:    "version without v prefix",
			version: "1.0.0",
			goos:    "linux",
			goarch:  "amd64",
			want:    "defradb_1.0.0_linux_x86_64",
			wantOk:  true,
		},
		{
			name:    "unsupported goos",
			version: "v1.0.0",
			goos:    "plan9",
			goarch:  "amd64",
			want:    "",
			wantOk:  false,
		},
		{
			name:    "unsupported goarch",
			version: "v1.0.0",
			goos:    "linux",
			goarch:  "riscv64",
			want:    "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := assetName(tt.version, tt.goos, tt.goarch)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// fakeDownloader records its calls and, unless failWith is set, writes a stub
// file to the requested destination.
type fakeDownloader struct {
	calls    int
	failWith error
}

func (f *fakeDownloader) download(ctx context.Context, tag string, asset string, destDir string) error {
	f.calls++
	if f.failWith != nil {
		return f.failWith
	}
	return os.WriteFile(filepath.Join(destDir, asset), []byte("stub"), 0o644)
}

func newTestProvider(t *testing.T, goos, goarch string) (*Provider, *fakeDownloader) {
	t.Helper()
	fd := &fakeDownloader{}
	p := &Provider{
		cacheDir: t.TempDir(),
		goos:     goos,
		goarch:   goarch,
		download: fd.download,
	}
	return p, fd
}

func TestBinaryPath_UnsupportedPlatform_Skips(t *testing.T) {
	p, fd := newTestProvider(t, "plan9", "amd64")

	path, skip, err := p.BinaryPath(context.Background(), "v1.0.0")

	require.NoError(t, err)
	assert.True(t, skip)
	assert.Empty(t, path)
	assert.Equal(t, 0, fd.calls)
}

func TestBinaryPath_CacheHit_SkipsDownload(t *testing.T) {
	p, fd := newTestProvider(t, "linux", "amd64")

	asset, ok := assetName("v1.0.0", "linux", "amd64")
	require.True(t, ok)
	cachedDir := filepath.Join(p.cacheDir, "v1.0.0")
	require.NoError(t, os.MkdirAll(cachedDir, 0o755))
	cachedPath := filepath.Join(cachedDir, asset)
	require.NoError(t, os.WriteFile(cachedPath, []byte("cached"), 0o755))

	path, skip, err := p.BinaryPath(context.Background(), "v1.0.0")

	require.NoError(t, err)
	assert.False(t, skip)
	assert.Equal(t, cachedPath, path)
	assert.Equal(t, 0, fd.calls)
}

func TestBinaryPath_CacheMiss_Downloads(t *testing.T) {
	p, fd := newTestProvider(t, "linux", "amd64")

	path, skip, err := p.BinaryPath(context.Background(), "v1.0.0")

	require.NoError(t, err)
	assert.False(t, skip)
	require.NotEmpty(t, path)
	assert.Equal(t, 1, fd.calls)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0o111, "expected downloaded binary to be executable")
}

func TestBinaryPath_DownloadError_IsHardError(t *testing.T) {
	p, fd := newTestProvider(t, "linux", "amd64")
	fd.failWith = errors.New("unexpected status downloading release asset: 404 Not Found")

	path, skip, err := p.BinaryPath(context.Background(), "v1.0.0")

	require.Error(t, err)
	assert.False(t, skip)
	assert.Empty(t, path)
	assert.ErrorContains(t, err, "404 Not Found")
	assert.Equal(t, 1, fd.calls)
}
