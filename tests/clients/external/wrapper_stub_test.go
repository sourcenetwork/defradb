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

package external

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubBinary builds testdata/stub once per test run and returns its path, so
// NewWrapper can be driven without execing a real defradb release binary.
var stubBinary = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "external-stub-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "stub")
	cmd := exec.Command("go", "build", "-o", path, "./testdata/stub")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	return path, nil
})

func requireStubBinary(t *testing.T) string {
	t.Helper()
	path, err := stubBinary()
	require.NoError(t, err, "failed to build test stub binary")
	return path
}

// TestNewWrapper_CtxCancelled_ReturnsPromptly starts a stub that never
// becomes healthy and gives NewWrapper a short-lived context. It asserts
// NewWrapper returns quickly (well inside the real healthCheckTimeout
// budget) once the context is cancelled, exercising the ctx-cancellation
// path in waitForHealth/NewWrapper without paying the full timeout cost.
func TestNewWrapper_CtxCancelled_ReturnsPromptly(t *testing.T) {
	binaryPath := requireStubBinary(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Setenv("STUB_MODE", "unhealthy")

	start := time.Now()
	w, err := NewWrapper(ctx, t, binaryPath)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, w)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, 10*time.Second, "NewWrapper should return promptly once ctx is cancelled")
}

// TestNewWrapper_StartFailure_ReturnsError points NewWrapper at a binary
// path that cannot be executed, exercising the cmd.Start() failure path.
func TestNewWrapper_StartFailure_ReturnsError(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "does-not-exist")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	w, err := NewWrapper(ctx, t, missingPath)

	require.Error(t, err)
	assert.Nil(t, w)
}

// TestWrapper_Close_Idempotent starts a healthy stub, gets a live Wrapper,
// and asserts calling Close twice does not panic.
func TestWrapper_Close_Idempotent(t *testing.T) {
	binaryPath := requireStubBinary(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	w, err := NewWrapper(ctx, t, binaryPath)
	require.NoError(t, err)
	require.NotNil(t, w)

	assert.NotPanics(t, func() {
		w.Close()
		w.Close()
	})
}
