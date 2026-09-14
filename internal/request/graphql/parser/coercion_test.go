// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package parser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateScalarRejectsOutOfRangeInt(t *testing.T) {
	require.ErrorContains(t, validateScalar("Int", float64(math.MaxInt32)+1),
		"Int cannot represent non 32-bit signed integer value")
}

func TestValidateScalarRejectsFractionalInt(t *testing.T) {
	require.ErrorContains(t, validateScalar("Int", 1.5),
		"Int cannot represent non 32-bit signed integer value")
}

func TestValidateScalarRejectsInvalidDateTime(t *testing.T) {
	require.EqualError(t, validateScalar("DateTime", "not-a-datetime"),
		`DateTime cannot represent value: "not-a-datetime"`)
}

func TestValidateScalarRejectsInvalidBlob(t *testing.T) {
	require.ErrorContains(t, validateScalar("Blob", "not-hex"), "Blob cannot represent value")
}

func TestValidateScalarAcceptsValidCustomScalars(t *testing.T) {
	require.NoError(t, validateScalar("Int", float64(math.MaxInt32)))
	require.NoError(t, validateScalar("DateTime", "2026-09-14T12:34:56.123Z"))
	require.NoError(t, validateScalar("DateTime", "UTC_NOW"))
	require.NoError(t, validateScalar("Blob", "0aB9"))
}
