// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithVersionAndCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			testUtils.Request{
				Request: `query {
					Users(cid: "bafyreic2xpowsfqw5vh42kjlyykrewjd77rsofsdfuz4slgvaeviv7hbbq") {
						Name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Chris",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
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
