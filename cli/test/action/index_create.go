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
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// IndexCreate executes the `client index create` command and requires that the returned
// result matches the expected value.
type IndexCreate struct {
	stateful
	augmented

	// The collection name to create the index on (required).
	Collection string

	// The name of the index (optional).
	// If not provided, a name will be generated automatically.
	Name string

	// The fields to index (required).
	// Format: "fieldName" or "fieldName:ASC" or "fieldName:DESC"
	Fields []string

	// Whether the index should be unique (optional, default: false).
	Unique bool

	// The expected index description.
	// If provided, it will be compared with the actual result.
	Expected immutable.Option[client.IndexDescription]

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*IndexCreate)(nil)

func (a *IndexCreate) Execute() {
	args := []string{"client", "index", "create"}

	if a.Collection != "" {
		args = append(args, "--collection", a.Collection)
	}

	if a.Name != "" {
		args = append(args, "--name", a.Name)
	}

	if len(a.Fields) > 0 {
		args = append(args, "--fields")
		fieldsStr := ""
		for i, field := range a.Fields {
			if i > 0 {
				fieldsStr += ","
			}
			fieldsStr += field
		}
		args = append(args, fieldsStr)
	}

	if a.Unique {
		args = append(args, "--unique")
	}

	args = append(args, a.AdditionalArgs...)
	args = a.AppendDirections(args)

	result, err := executeJson[client.IndexDescription](a.s.Ctx, args)

	if a.ExpectError != "" {
		require.Error(a.s.T, err)
		require.Contains(a.s.T, err.Error(), a.ExpectError)
		return
	}

	require.NoError(a.s.T, err)

	if a.Expected.HasValue() {
		expected := a.Expected.Value()
		if expected.ID != 0 {
			require.Equal(a.s.T, expected.ID, result.ID)
		}
		if expected.Name != "" {
			require.Equal(a.s.T, expected.Name, result.Name)
		}
		require.Equal(a.s.T, expected.Fields, result.Fields)
		require.Equal(a.s.T, expected.Unique, result.Unique)
	}
}
