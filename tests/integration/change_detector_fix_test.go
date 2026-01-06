// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/change_detector"
)

// TestGetActionRange_WithCreateDoc_SplitsBeforeCreateDoc verifies that the change detector
// correctly splits BEFORE CreateDoc actions.
//
// CreateDoc actions should NOT be treated as setup actions to ensure that both the source
// and target branches execute document creation operations, enabling detection of
// breaking changes in write operations.
func TestGetActionRange_WithCreateDoc_SplitsBeforeCreateDoc(t *testing.T) {
	originalEnabled := change_detector.Enabled
	originalSetupOnly := change_detector.SetupOnly

	os.Setenv("DEFRA_CHANGE_DETECTOR_ENABLE", "true")
	change_detector.Enabled = true
	change_detector.SetupOnly = false

	defer func() {
		change_detector.Enabled = originalEnabled
		change_detector.SetupOnly = originalSetupOnly
		os.Unsetenv("DEFRA_CHANGE_DETECTOR_ENABLE")
	}()

	schema := &action.AddSchema{
		Schema: `
			type User {
				name: String
			}
		`,
	}

	testCase := TestCase{
		Actions: []any{
			schema,
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Jane"}`,
			},
			Request{
				Request: `query { User { name } }`,
			},
		},
	}

	startIndex, endIndex := getActionRange(t, testCase)

	// The split should happen BEFORE the first CreateDoc (index 1).
	// Target branch starts at index 1 (first CreateDoc) and ends at index 3 (Request).
	assert.Equal(t, 1, startIndex, "startIndex should be at the first CreateDoc action (index 1)")
	assert.Equal(t, 3, endIndex, "endIndex should be at the last action (Request)")
}

// TestGetActionRange_WithCreateDoc_SetupModeEndsBeforeCreateDoc verifies that in setup mode,
// the change detector ends BEFORE the first CreateDoc action.
//
// This ensures that the source branch only executes schema setup, while document creation
// is left for both branches to execute.
func TestGetActionRange_WithCreateDoc_SetupModeEndsBeforeCreateDoc(t *testing.T) {
	originalEnabled := change_detector.Enabled
	originalSetupOnly := change_detector.SetupOnly

	os.Setenv("DEFRA_CHANGE_DETECTOR_ENABLE", "true")
	os.Setenv("DEFRA_CHANGE_DETECTOR_SETUP_ONLY", "true")
	change_detector.Enabled = true
	change_detector.SetupOnly = true

	defer func() {
		change_detector.Enabled = originalEnabled
		change_detector.SetupOnly = originalSetupOnly
		os.Unsetenv("DEFRA_CHANGE_DETECTOR_ENABLE")
		os.Unsetenv("DEFRA_CHANGE_DETECTOR_SETUP_ONLY")
	}()

	schema := &action.AddSchema{
		Schema: `
			type User {
				name: String
			}
		`,
	}

	testCase := TestCase{
		Actions: []any{
			schema,
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Jane"}`,
			},
			Request{
				Request: `query { User { name } }`,
			},
		},
	}

	startIndex, endIndex := getActionRange(t, testCase)

	// In setup mode, endIndex should be 0 (only schema) since CreateDoc is not a setup action.
	assert.Equal(t, 0, startIndex, "startIndex should be 0 in setup mode")
	assert.Equal(t, 0, endIndex, "endIndex should be 0 (only AddSchema)")
}

// TestGetActionRange_WithSetupComplete_RespectExplicitSplit verifies that explicit
// SetupComplete markers placed after schema actions are respected.
//
// Note: If SetupComplete is placed after CreateDoc, the split will happen at CreateDoc first
// because CreateDoc triggers the split.
func TestGetActionRange_WithSetupComplete_RespectExplicitSplit(t *testing.T) {
	originalEnabled := change_detector.Enabled
	originalSetupOnly := change_detector.SetupOnly

	os.Setenv("DEFRA_CHANGE_DETECTOR_ENABLE", "true")
	change_detector.Enabled = true
	change_detector.SetupOnly = false

	defer func() {
		change_detector.Enabled = originalEnabled
		change_detector.SetupOnly = originalSetupOnly
		os.Unsetenv("DEFRA_CHANGE_DETECTOR_ENABLE")
	}()

	schema := &action.AddSchema{
		Schema: `
			type User {
				name: String
			}
		`,
	}

	// SetupComplete is placed AFTER schema, BEFORE any CreateDoc.
	// This is the correct way to use SetupComplete if you want schema-only setup.
	testCase := TestCase{
		Actions: []any{
			schema,
			SetupComplete{},
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Jane"}`,
			},
			Request{
				Request: `query { User { name } }`,
			},
		},
	}

	startIndex, endIndex := getActionRange(t, testCase)

	// With explicit SetupComplete at index 1, split should happen there.
	// startIndex should be 2 (action after SetupComplete).
	assert.Equal(t, 2, startIndex, "startIndex should be 2 (after SetupComplete)")
	assert.Equal(t, 4, endIndex, "endIndex should be 4 (last action)")
}

// TestGetActionRange_CreateDocWithExpectedError_ExecutedOnTargetBranch verifies that
// CreateDoc actions with ExpectedError are executed on the target branch.
//
// This is critical for detecting breaking changes in write operations. For example,
// if a test expects an error when creating a duplicate relationship, that error
// validation must happen on the target branch to catch any regressions.
func TestGetActionRange_CreateDocWithExpectedError_ExecutedOnTargetBranch(t *testing.T) {
	originalEnabled := change_detector.Enabled
	originalSetupOnly := change_detector.SetupOnly

	os.Setenv("DEFRA_CHANGE_DETECTOR_ENABLE", "true")
	change_detector.Enabled = true
	change_detector.SetupOnly = false

	defer func() {
		change_detector.Enabled = originalEnabled
		change_detector.SetupOnly = originalSetupOnly
		os.Unsetenv("DEFRA_CHANGE_DETECTOR_ENABLE")
	}()

	schema := &action.AddSchema{
		Schema: `
			type Book {
				title: String
			}
			type Author {
				name: String
				book: Book @primary
			}
		`,
	}

	testCase := TestCase{
		Actions: []any{
			schema,
			CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book1"}`,
			},
			CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "Author1"}`,
			},
			// This CreateDoc expects an error - it must be executed on target branch
			// to verify the error is still raised correctly.
			CreateDoc{
				CollectionID:  1,
				Doc:           `{"name": "Author2"}`,
				ExpectedError: "already linked",
			},
		},
	}

	startIndex, endIndex := getActionRange(t, testCase)

	// startIndex should be 1 (first CreateDoc), meaning the target branch WILL
	// execute CreateDoc actions, including the one with ExpectedError.
	assert.Equal(t, 1, startIndex,
		"startIndex should be 1 (first CreateDoc), ensuring error validation runs on target branch")
	assert.Equal(t, 3, endIndex, "endIndex should be 3 (last action)")
}
