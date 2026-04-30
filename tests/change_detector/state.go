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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// stateFileName is the filename used for the per-test change-detector
// sidecar inside DatabaseDir(t).
const stateFileName = "_change_detector_state.json"

// File mode for the per-test data directory. Standard test-fixture default
// (rwx for owner, rx for everyone else); matches what t.TempDir and badger
// use, no security-sensitive content.
const stateDirMode = 0o755

// File mode for the sidecar JSON file. Standard test-fixture default
// (rw for owner, r for everyone else); no security-sensitive content.
const stateFileMode = 0o644

// TestState is the slice of in-memory test harness state that the source
// phase of the change detector hands to the assert phase via a JSON sidecar
// in the per-test data directory.
//
// The struct is JSON-serialized; new fields can be added in a forward and
// backward compatible way because encoding/json silently ignores unknown
// fields on read and zero-values missing fields.
type TestState struct {
	// CollectionVersions is the ordered list of collection version IDs created
	// during the source phase, used to resolve {{.CollectionVersionIDN}}
	// templates on the assert side against the values the source side produced.
	CollectionVersions []string `json:"collectionVersions"`
}

// stateFilePath returns the absolute path of the sidecar file for the given test.
func stateFilePath(t testing.TB) string {
	return filepath.Join(DatabaseDir(t), stateFileName)
}

// WriteTestState marshals state as JSON and writes it into the per-test
// change-detector data directory. Callers must guard with SetupOnly.
func WriteTestState(t testing.TB, state TestState) error {
	if err := os.MkdirAll(DatabaseDir(t), stateDirMode); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFilePath(t), data, stateFileMode)
}

// ReadTestState reads the sidecar file written by the matching source-phase
// invocation. Callers can detect a missing file via errors.Is(err, fs.ErrNotExist)
// to fall back gracefully when the source branch pre-dates this mechanism.
func ReadTestState(t testing.TB) (TestState, error) {
	data, err := os.ReadFile(stateFilePath(t))
	if err != nil {
		return TestState{}, err
	}

	var state TestState
	return state, json.Unmarshal(data, &state)
}
