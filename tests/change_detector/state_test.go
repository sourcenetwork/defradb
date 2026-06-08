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

package change_detector

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// setRootDatabaseDir sets rootDatabaseDir to dir for the duration of the test,
// restoring the original value via t.Cleanup.
func setRootDatabaseDir(t *testing.T, dir string) {
	t.Helper()
	old := rootDatabaseDir
	rootDatabaseDir = dir
	t.Cleanup(func() { rootDatabaseDir = old })
}

func TestWriteReadTestState_RoundTrip(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	want := TestState{
		CollectionVersions: []string{"v0", "v1"},
	}

	if err := WriteTestState(t, want); err != nil {
		t.Fatalf("WriteTestState: %v", err)
	}

	got, err := ReadTestState(t)
	if err != nil {
		t.Fatalf("ReadTestState error: %v", err)
	}

	if len(got.CollectionVersions) != len(want.CollectionVersions) {
		t.Fatalf("CollectionVersions len: got %d, want %d", len(got.CollectionVersions), len(want.CollectionVersions))
	}
	for i, v := range want.CollectionVersions {
		if got.CollectionVersions[i] != v {
			t.Errorf("CollectionVersions[%d]: got %q, want %q", i, got.CollectionVersions[i], v)
		}
	}
}

func TestReadTestState_MissingFile(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	_, err := ReadTestState(t)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadTestState: expected fs.ErrNotExist, got %v", err)
	}
}

func TestReadTestState_CorruptFile(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	dir := DatabaseDir(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(stateFilePath(t), []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := ReadTestState(t)
	if err == nil {
		t.Fatal("ReadTestState: expected non-nil error on corrupt file, got nil")
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadTestState: corrupt file should not surface as ErrNotExist, got %v", err)
	}
}

func TestWriteReadTestState_EmptyState(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	if err := WriteTestState(t, TestState{}); err != nil {
		t.Fatalf("WriteTestState: %v", err)
	}

	got, err := ReadTestState(t)
	if err != nil {
		t.Fatalf("ReadTestState error: %v", err)
	}

	if len(got.CollectionVersions) != 0 {
		t.Errorf("CollectionVersions: expected empty, got %v", got.CollectionVersions)
	}

	expectedPath := filepath.Join(DatabaseDir(t), stateFileName)
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("expected file at %s, got: %v", expectedPath, err)
	}
}
