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

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	defraMultiplier "github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestApplyTestCaseLevelMultipliers_WithSignedDocs_EnablesSigning(t *testing.T) {
	tc := &TestCase{EnableSigning: false}

	applyTestCaseLevelMultipliers(tc, defraMultiplier.SignedDocs)

	assert.True(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_WithSignedDocsAlreadyTrue_RemainsTrue(t *testing.T) {
	tc := &TestCase{EnableSigning: true}

	applyTestCaseLevelMultipliers(tc, defraMultiplier.SignedDocs)

	assert.True(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_WithoutSignedDocs_LeavesFlagUnchanged(t *testing.T) {
	tc := &TestCase{EnableSigning: false}

	// Intentionally passes an unrelated multiplier name. The hook must
	// be a no-op for anything it does not recognize.
	applyTestCaseLevelMultipliers(tc, "secondary-index")

	assert.False(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_WithEmptyActiveSet_LeavesFlagUnchanged(t *testing.T) {
	tc := &TestCase{EnableSigning: false}

	applyTestCaseLevelMultipliers(tc, "")

	assert.False(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_WithMultipleMultipliersIncludingSignedDocs_EnablesSigning(t *testing.T) {
	tc := &TestCase{EnableSigning: false}

	applyTestCaseLevelMultipliers(tc, "secondary-index,signed-docs")

	assert.True(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_IgnoresSurroundingWhitespace(t *testing.T) {
	tc := &TestCase{EnableSigning: false}

	// testo's Init() call TrimSpaces names, but Get() echoes whatever
	// was stored. Our hook uses TrimSpace defensively — verify that.
	applyTestCaseLevelMultipliers(tc, "  signed-docs  ")

	assert.True(t, tc.EnableSigning)
}

func TestApplyTestCaseLevelMultipliers_NeverDowngradesEnableSigning(t *testing.T) {
	// Lock in the no-downgrade contract. Even an empty multiplier set
	// must not turn EnableSigning back to false.
	tc := &TestCase{EnableSigning: true}

	applyTestCaseLevelMultipliers(tc, "")
	assert.True(t, tc.EnableSigning, "empty multiplier set must not downgrade")

	applyTestCaseLevelMultipliers(tc, "signed-docs")
	assert.True(t, tc.EnableSigning, "signed-docs must not downgrade an already-true flag")

	applyTestCaseLevelMultipliers(tc, "secondary-index")
	assert.True(t, tc.EnableSigning, "unrelated multipliers must not touch the flag")
}
