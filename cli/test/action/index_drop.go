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

import (
	"github.com/stretchr/testify/require"
)

// IndexDrop executes the `client index drop` command.
type IndexDrop struct {
	stateful
	augmented

	// The collection name containing the index (required).
	Collection string

	// The name of the index to drop (required).
	Name string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*IndexDrop)(nil)

func (a *IndexDrop) Execute() {
	args := []string{"client", "index", "drop"}

	if a.Collection != "" {
		args = append(args, "--collection", a.Collection)
	}

	if a.Name != "" {
		args = append(args, "--name", a.Name)
	}

	args = append(args, a.AdditionalArgs...)
	args = a.AppendDirections(args)

	err := execute(a.s.Ctx, args)

	if a.ExpectError != "" {
		require.Error(a.s.T, err)
		require.Contains(a.s.T, err.Error(), a.ExpectError)
		return
	}

	require.NoError(a.s.T, err)
}
