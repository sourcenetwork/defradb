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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// stateFileName is the filename used for the per-test change-detector
// sidecar inside DatabaseDir(t).
const stateFileName = "_change_detector_state.json"

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
	// LensIDs is the ordered list of lens IDs registered during the source
	// phase. Reserved for future use; not currently consumed by any test.
	LensIDs []string `json:"lensIDs,omitempty"`
}

// stateFilePath returns the absolute path of the sidecar file for the given test.
func stateFilePath(t testing.TB) string {
	return filepath.Join(DatabaseDir(t), stateFileName)
}

// WriteTestState marshals state as JSON and writes it atomically into the
// per-test change-detector data directory. The write goes to a sibling
// `*.tmp` file and is then renamed into place, so a partially-written file
// is never visible to a concurrent reader (or to the assert-phase process).
//
// It is a programming error to call this when changeDetector is not in
// SetupOnly mode; callers must guard accordingly.
func WriteTestState(t testing.TB, state TestState) error {
	dir := DatabaseDir(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	tmpPath := stateFilePath(t) + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, stateFilePath(t)); err != nil {
		// best-effort cleanup of the tmp file on failure
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// ReadTestState reads the sidecar file written by WriteTestState in the
// matching source-phase test invocation. The bool return distinguishes
// "file missing" (acceptable: source branch may pre-date the sidecar, or
// the test wrote nothing) from "file corrupt" (a real error).
//
//   - (state, true,  nil) — file existed and parsed cleanly
//   - (zero,  false, nil) — file did not exist
//   - (_,     false, err) — file existed but could not be read or parsed
func ReadTestState(t testing.TB) (TestState, bool, error) {
	data, err := os.ReadFile(stateFilePath(t))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return TestState{}, false, nil
		}
		return TestState{}, false, err
	}

	var state TestState
	if err := json.Unmarshal(data, &state); err != nil {
		return TestState{}, false, err
	}
	return state, true, nil
}
