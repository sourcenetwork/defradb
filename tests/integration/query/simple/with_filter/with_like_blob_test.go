// Copyright 2024 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithLikeOpOnBlobField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						data: Blob
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"data": "00FF99AA"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"data": "FA02CC45"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {data: {_like: "%FF99%"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
