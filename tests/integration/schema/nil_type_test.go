// Copyright 2024 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchema_WithMissingType_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name:
					}
				`,
			},
		},
	}

	require.Panics(
		t,
		func() {
			testUtils.ExecuteTestCase(t, test)
		},
	)
}
