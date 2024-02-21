// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_ReturnsVersionCID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with version cid",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {name: "John"}) {
								_version {
									cid
								}
							}
						}`,
				Results: []map[string]any{
					{
						"_version": []map[string]any{
							{
								"cid": "bafybeidzstxabh7qktq7pkmmxvpjbnwklxz3h5l6d425ldvjy65xvvuxu4",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
