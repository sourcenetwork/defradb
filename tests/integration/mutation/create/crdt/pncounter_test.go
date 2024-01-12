// Copyright 2023 Democratized Data Foundation
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

func TestPNCounterCreate_IntKindWithPositiveValue_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Document creation with PN Counter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						_docID
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-a688789e-d8a6-57a7-be09-22e005ab79e0",
						"name":   "John",
						"points": int64(10),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
