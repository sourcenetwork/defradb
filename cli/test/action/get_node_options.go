// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/stretchr/testify/require"
)

// NodeOptionExpected describes a single field assertion against the node-options JSON.
type NodeOptionExpected struct {
	// Path is the sequence of JSON keys to traverse, e.g. []string{"HTTP", "Address"}.
	Path []string
	// Value is the expected value. JSON numbers unmarshal as float64; booleans as bool;
	// arrays as []any; null as nil.
	// Ignored when UseRootDir is true.
	Value any
	// UseRootDir replaces Value with the test's RootDir at assertion time. Use this for
	// path fields whose value is set to the node's root directory (e.g. DocumentACP.Path).
	UseRootDir bool
}

// AssertNodeOptions executes `client node-options` and asserts that specific fields
// in the returned JSON match the expected values.
type AssertNodeOptions struct {
	stateful
	Expected []NodeOptionExpected
}

var _ Action = (*AssertNodeOptions)(nil)

func (a *AssertNodeOptions) Execute() {
	args := []string{"client", "node-options"}
	args = a.AppendDirections(args)

	result, err := executeJson[map[string]any](a.s.Ctx, args)
	require.NoError(a.s.T, err)

	for _, exp := range a.Expected {
		actual, ok := navigateOptionsMap(result, exp.Path...)
		require.True(a.s.T, ok, "field path %v not found in node options", exp.Path)

		expected := exp.Value
		if exp.UseRootDir {
			expected = a.s.RootDir
		}
		require.Equal(a.s.T, expected, actual, "node options field %v", exp.Path)
	}
}

// navigateOptionsMap traverses a nested map[string]any by following keys in order,
// returning the value at the final key and true, or nil and false if any key is
// missing or an intermediate value is not itself a map.
func navigateOptionsMap(m map[string]any, keys ...string) (any, bool) {
	cur := any(m)
	for _, k := range keys {
		sub, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = sub[k]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
