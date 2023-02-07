// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldWithCreateWithUpdateAfterSchemaUpdateAndVersionJoin(t *testing.T) {
	initialSchemaVersionId := "bafkreicg3xcpjlt3ecguykpcjrdx5ogi4n7cq2fultyr6vippqdxnrny3u"
	updatedSchemaVersionId := "bafkreicnj2kiq6vqxozxnhrc4mlbkdp5rr44awaetn5x5hcdymk6lxrxdy"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with update after schema update, verison join",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John"
				}`,
			},
			// We want to make sure that this works across database versions, so we tell
			// the change detector to split here.
			testUtils.Request{
				Request: `query {
					Users {
						Name
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "John",
						"_version": []map[string]any{
							{
								"schemaVersionId": initialSchemaVersionId,
							},
						},
					},
				},
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: []map[string]any{
					{
						"Name":  "John",
						"Email": "ih8oraclelicensing@netscape.net",
						"_version": []map[string]any{
							{
								// Update commit
								"schemaVersionId": updatedSchemaVersionId,
							},
							{
								// Create commit
								"schemaVersionId": initialSchemaVersionId,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldWithCreateWithUpdateAfterSchemaUpdateAndCommitQuery(t *testing.T) {
	initialSchemaVersionId := "bafkreicg3xcpjlt3ecguykpcjrdx5ogi4n7cq2fultyr6vippqdxnrny3u"
	updatedSchemaVersionId := "bafkreicnj2kiq6vqxozxnhrc4mlbkdp5rr44awaetn5x5hcdymk6lxrxdy"

	test := testUtils.TestCase{
		Description: "Test schema update, add field with update after schema update, commits query",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			testUtils.Request{
				Request: `query {
					commits (field: "C") {
						schemaVersionId
					}
				}`,
				Results: []map[string]any{
					{
						// Update commit
						"schemaVersionId": updatedSchemaVersionId,
					},
					{
						// Create commit
						"schemaVersionId": initialSchemaVersionId,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
