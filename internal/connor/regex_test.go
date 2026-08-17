// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package connor

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"
)

func TestRegex(t *testing.T) {
	const testString = "Source is the glue of web3"

	// match anywhere
	result, err := regex("glue", testString)
	require.NoError(t, err)
	require.True(t, result)

	// no match
	result, err = regex("gluey", testString)
	require.NoError(t, err)
	require.False(t, result)

	// anchored at both ends
	result, err = regex("^Source.*web3$", testString)
	require.NoError(t, err)
	require.True(t, result)

	// anchored at the start, no match
	result, err = regex("^glue", testString)
	require.NoError(t, err)
	require.False(t, result)

	// character class
	result, err = regex("web[0-9]", testString)
	require.NoError(t, err)
	require.True(t, result)

	// alternation
	result, err = regex("(paste|glue)", testString)
	require.NoError(t, err)
	require.True(t, result)

	// regular expressions are case sensitive
	result, err = regex("SOURCE", testString)
	require.NoError(t, err)
	require.False(t, result)

	// a case insensitive match is expressed in the pattern
	result, err = regex("(?i)SOURCE", testString)
	require.NoError(t, err)
	require.True(t, result)
}

func TestRegex_WithNillableString(t *testing.T) {
	result, err := regex("glue", immutable.Some("the glue of web3"))
	require.NoError(t, err)
	require.True(t, result)

	// a nil value only matches a nil condition
	result, err = regex("glue", immutable.None[string]())
	require.NoError(t, err)
	require.False(t, result)
}

func TestRegex_WithNonStringData_DoesNotMatch(t *testing.T) {
	result, err := regex("3", 3)
	require.NoError(t, err)
	require.False(t, result)
}

func TestRegex_WithInvalidPattern_ReturnsError(t *testing.T) {
	_, err := regex("web[0-9", "Source is the glue of web3")
	require.ErrorIs(t, err, ErrInvalidRegex)
}

func TestRegex_WithNonStringCondition_ReturnsError(t *testing.T) {
	_, err := regex(3, "Source is the glue of web3")
	require.Error(t, err)
}
