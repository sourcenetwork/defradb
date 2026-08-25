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

package action

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// areResultsEqual used to compare maps by length alone and then index the
// actual map directly, so a missing key resolved to a nil zero value and
// compared equal to an expected nil. Two maps with entirely different keys
// could therefore be reported as equal. This is a regression test for the fix.
func TestAreResultsEqual_MapWithDifferentKeys_DoesNotMatch(t *testing.T) {
	expected := map[string]any{"a": nil}
	actual := map[string]any{"b": nil}

	assert.False(t, areResultsEqual(expected, actual))
}

func TestAreResultsEqual_MapWithSameKeyAndNilValue_Matches(t *testing.T) {
	expected := map[string]any{"a": nil}
	actual := map[string]any{"a": nil}

	assert.True(t, areResultsEqual(expected, actual))
}

func TestAreResultsEqual_Uint64AboveMaxInt64_Matches(t *testing.T) {
	expected := uint64(math.MaxInt64) + 1
	actual := json.Number(strconv.FormatUint(expected, 10))

	assert.True(t, areResultsEqual(expected, actual))
}

func TestAreResultsEqual_MaxUint64_Matches(t *testing.T) {
	expected := uint64(math.MaxUint64)
	actual := json.Number(strconv.FormatUint(expected, 10))

	assert.True(t, areResultsEqual(expected, actual))
}

// Parsing a negative JSON number with Int64() and then letting
// assert.ObjectsAreEqualValues convert the signed result back to a uint64
// wrapped -1 around to math.MaxUint64, so the two incorrectly compared equal.
// This is a regression test of the fix.
func TestAreResultsEqual_MaxUint64AgainstNegativeOne_DoesNotMatch(t *testing.T) {
	expected := uint64(math.MaxUint64)
	actual := json.Number("-1")

	assert.False(t, areResultsEqual(expected, actual))
}
