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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withReleaseBaseURL points releaseBaseURL at server for the duration of the
// test, restoring the original value on cleanup.
func withReleaseBaseURL(t *testing.T, server *httptest.Server) {
	t.Helper()
	original := releaseBaseURL
	releaseBaseURL = server.URL
	t.Cleanup(func() {
		releaseBaseURL = original
	})
}

// assertOnlyFile asserts destDir contains exactly the given files (order
// independent), so tests can confirm no leftover *.tmp files remain.
func assertOnlyFiles(t *testing.T, destDir string, want []string) {
	t.Helper()
	entries, err := os.ReadDir(destDir)
	require.NoError(t, err)
	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.Name())
	}
	assert.ElementsMatch(t, want, got)
}

func TestHttpDownload_Success(t *testing.T) {
	const asset = "defradb_1.0.0_linux_x86_64"
	const body = "fake release binary bytes"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/"+repo+"/releases/download/v1.0.0/"+asset, r.URL.Path)
		_, err := w.Write([]byte(body))
		assert.NoError(t, err)
	}))
	defer server.Close()
	withReleaseBaseURL(t, server)

	destDir := t.TempDir()
	err := httpDownload(context.Background(), "v1.0.0", asset, destDir)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(destDir, asset))
	require.NoError(t, err)
	assert.Equal(t, body, string(got))

	assertOnlyFiles(t, destDir, []string{asset})
}

func TestHttpDownload_NotFound_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()
	withReleaseBaseURL(t, server)

	destDir := t.TempDir()
	err := httpDownload(context.Background(), "v1.0.0", "defradb_1.0.0_linux_x86_64", destDir)

	require.Error(t, err)
	assert.ErrorContains(t, err, "404")

	assertOnlyFiles(t, destDir, []string{})
}

func TestHttpDownload_Redirect_Follows(t *testing.T) {
	const asset = "defradb_1.0.0_linux_x86_64"
	const body = "fake release binary bytes"

	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(body))
		assert.NoError(t, err)
	}))
	defer final.Close()

	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+r.URL.Path, http.StatusFound)
	}))
	defer redirecting.Close()
	withReleaseBaseURL(t, redirecting)

	destDir := t.TempDir()
	err := httpDownload(context.Background(), "v1.0.0", asset, destDir)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(destDir, asset))
	require.NoError(t, err)
	assert.Equal(t, body, string(got))

	assertOnlyFiles(t, destDir, []string{asset})
}

// TestHttpDownload_BodyCopyFailure_CleansUpTempFile forces io.Copy to fail
// deterministically by advertising a Content-Length larger than the bytes
// actually sent, then hanging up the connection. This is deterministic
// (unlike e.g. a timing-based hang up) because the client always sees fewer
// bytes than promised and net/http always surfaces that as ErrUnexpectedEOF.
func TestHttpDownload_BodyCopyFailure_CleansUpTempFile(t *testing.T) {
	const asset = "defradb_1.0.0_linux_x86_64"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("short"))
		assert.NoError(t, err)

		hj, ok := w.(http.Hijacker)
		require.True(t, ok)
		conn, _, err := hj.Hijack()
		require.NoError(t, err)
		_ = conn.Close()
	}))
	defer server.Close()
	withReleaseBaseURL(t, server)

	destDir := t.TempDir()
	err := httpDownload(context.Background(), "v1.0.0", asset, destDir)

	require.Error(t, err)
	assertOnlyFiles(t, destDir, []string{})
}
