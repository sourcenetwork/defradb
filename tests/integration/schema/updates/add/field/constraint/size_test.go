// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdates_AddFieldSizeContraint_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add size contraint to array field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 9, "Typ":1, "Size": 2} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(true),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
					"foo":  []float32{1, 2, 3},
				},
				// ExpectedError: ,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
