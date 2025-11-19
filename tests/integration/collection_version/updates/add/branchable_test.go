// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBranchableCollection_AddNewField_ShouldAddField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			testUtils.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name":  "Andy",
					"email": "andy@gmail.com",
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  1,
				Doc: `{
					"email": "islam@gmail.com"
				}`,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					User {
						name
						email
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "John",
							"email": nil,
						},
						{
							"name":  "Islam",
							"email": "islam@gmail.com",
						},
						{
							"name":  "Andy",
							"email": "andy@gmail.com",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
