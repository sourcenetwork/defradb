// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSDL_WithNestedListField_ReturnsError(t *testing.T) {
	schemaManager, err := NewSchemaManager(false)
	require.NoError(t, err)

	_, err = schemaManager.ParseSDL(`
		type User {
			name: [[String]]
		}
	`)
	require.ErrorContains(t, err, errNestedListTypeNotSupported)
	require.ErrorContains(t, err, "User")
	require.ErrorContains(t, err, "name")
}
