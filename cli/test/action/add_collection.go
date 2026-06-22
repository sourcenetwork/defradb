// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import "github.com/stretchr/testify/require"

// AddCollection executes the `client collection add` command using the given SDL.
type AddCollection struct {
	stateful
	augmented

	// The SDL string value to be passed directly to the command (i.e. not via a file)
	InlineSDL string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*AddCollection)(nil)

func (a *AddCollection) Execute() {
	args := []string{"client", "collection", "add"}

	args = append(args, a.InlineSDL)

	args = a.AppendDirections(args)
	args = append(args, a.AdditionalArgs...)

	err := execute(a.s.Ctx, args)

	if a.ExpectError != "" {
		require.Error(a.s.T, err)
		require.Contains(a.s.T, err.Error(), a.ExpectError)
		return
	}

	require.NoError(a.s.T, err)
}
