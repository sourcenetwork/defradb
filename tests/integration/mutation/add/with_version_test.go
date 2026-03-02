// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAdd_ReturnsVersionCID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
							add_Users(input: {name: "John"}) {
								_version {
									cid
								}
							}
						}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"_version": []map[string]any{
								{
									"cid": "bafyreifldhofx6cwi6ashk24rcefsuiqje5a2rziwcyte54z27wmgv4pey",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
