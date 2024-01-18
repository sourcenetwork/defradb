// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithDateTimeField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update of date time field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"created_at": "2021-07-23T02:22:22-05:00"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							created_at
						}
					}
				`,
				Results: []map[string]any{
					{
						"created_at": testUtils.MustParseTime("2021-07-23T02:22:22-05:00"),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_WithDateTimeField_MultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update of date time field, multiple docs",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"created_at": "2021-07-23T02:22:22-05:00"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(input: {created_at: "2031-07-23T03:23:23Z"}) {
						name
						created_at
					}
				}`,
				Results: []map[string]any{
					{
						"name":       "Fred",
						"created_at": testUtils.MustParseTime("2031-07-23T03:23:23Z"),
					},
					{
						"name":       "John",
						"created_at": testUtils.MustParseTime("2031-07-23T03:23:23Z"),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
