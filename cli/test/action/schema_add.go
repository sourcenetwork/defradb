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

// SchemaAdd executes the `client schema add` command using the given schema.
type SchemaAdd struct {
	stateful
	argmented

	// The schema string value to be passed directly to the command (i.e. not via a file)
	InlineSchema string
}

var _ Action = (*SchemaAdd)(nil)

func (a *SchemaAdd) Execute() {
	args := []string{"client", "schema", "add"}

	args = append(args, a.InlineSchema)

	args = a.AppendDirections(args)
	args = append(args, a.AdditionalArgs...)

	err := execute(a.s.Ctx, args)
	require.NoError(a.s.T, err)
}
