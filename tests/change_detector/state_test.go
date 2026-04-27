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
		LensIDs:            []string{"l0"},
	}

	if err := WriteTestState(t, want); err != nil {
		t.Fatalf("WriteTestState: %v", err)
	}

	got, found, err := ReadTestState(t)
	if err != nil {
		t.Fatalf("ReadTestState error: %v", err)
	}
	if !found {
		t.Fatal("ReadTestState: expected found=true, got false")
	}

	if len(got.CollectionVersions) != len(want.CollectionVersions) {
		t.Fatalf("CollectionVersions len: got %d, want %d", len(got.CollectionVersions), len(want.CollectionVersions))
	}
	for i, v := range want.CollectionVersions {
		if got.CollectionVersions[i] != v {
			t.Errorf("CollectionVersions[%d]: got %q, want %q", i, got.CollectionVersions[i], v)
		}
	}

	if len(got.LensIDs) != len(want.LensIDs) {
		t.Fatalf("LensIDs len: got %d, want %d", len(got.LensIDs), len(want.LensIDs))
	}
	for i, v := range want.LensIDs {
		if got.LensIDs[i] != v {
			t.Errorf("LensIDs[%d]: got %q, want %q", i, got.LensIDs[i], v)
		}
	}
}

func TestReadTestState_MissingFile(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	got, found, err := ReadTestState(t)
	if err != nil {
		t.Fatalf("ReadTestState error: %v", err)
	}
	if found {
		t.Fatal("ReadTestState: expected found=false, got true")
	}
	if len(got.CollectionVersions) != 0 || len(got.LensIDs) != 0 {
		t.Errorf("ReadTestState: expected zero-value state, got %+v", got)
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

	_, found, err := ReadTestState(t)
	if found {
		t.Fatal("ReadTestState: expected found=false on corrupt file, got true")
	}
	if err == nil {
		t.Fatal("ReadTestState: expected non-nil error on corrupt file, got nil")
	}
}

func TestWriteTestState_NoLeftoverTmp(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	state := TestState{CollectionVersions: []string{"v0"}}
	if err := WriteTestState(t, state); err != nil {
		t.Fatalf("WriteTestState: %v", err)
	}

	tmpPath := stateFilePath(t) + ".tmp"
	_, err := os.Stat(tmpPath)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected .tmp file to be gone after WriteTestState, but stat returned: %v", err)
	}

	// Verify the final file exists
	if _, err := os.Stat(stateFilePath(t)); err != nil {
		t.Errorf("expected state file to exist, got: %v", err)
	}
}

func TestWriteReadTestState_EmptyState(t *testing.T) {
	setRootDatabaseDir(t, t.TempDir())

	if err := WriteTestState(t, TestState{}); err != nil {
		t.Fatalf("WriteTestState: %v", err)
	}

	got, found, err := ReadTestState(t)
	if err != nil {
		t.Fatalf("ReadTestState error: %v", err)
	}
	if !found {
		t.Fatal("ReadTestState: expected found=true, got false")
	}

	if len(got.CollectionVersions) != 0 {
		t.Errorf("CollectionVersions: expected empty, got %v", got.CollectionVersions)
	}
	if len(got.LensIDs) != 0 {
		t.Errorf("LensIDs: expected empty, got %v", got.LensIDs)
	}

	// Verify stateFilePath helper indirectly by ensuring the file lives in DatabaseDir
	expectedDir := DatabaseDir(t)
	expectedPath := filepath.Join(expectedDir, stateFileName)
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("expected file at %s, got: %v", expectedPath, err)
	}
}
