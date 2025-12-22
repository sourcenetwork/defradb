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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldWithCreateWithUpdateAfterSchemaUpdateAndVersionJoin(t *testing.T) {
	initialSchemaVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	updatedSchemaVersionID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			// We want to make sure that this works across database versions, so we tell
			// the change detector to split here.
			testUtils.Request{
				Request: `query {
					Users {
						name
						_version {
							schemaVersionId
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"schemaVersionId": initialSchemaVersionID,
									"fieldName":       "_C",
								},
							},
						},
					},
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"email": "ih8oraclelicensing@netscape.net",
							"_version": []map[string]any{
								{
									// Update commit
									"schemaVersionId": updatedSchemaVersionID,
								},
								{
									// Create commit
									"schemaVersionId": initialSchemaVersionID,
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

func TestSchemaUpdatesAddFieldWithCreateWithUpdateAfterSchemaUpdateAndCommitQuery(t *testing.T) {
	initialSchemaVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	updatedSchemaVersionID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"email": "ih8oraclelicensing@netscape.net"
				}`,
			},
			testUtils.Request{
				Request: `query {
					_commits (filter: {fieldName: {_eq: "_C"}}) {
						schemaVersionId
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// Update commit
							"schemaVersionId": updatedSchemaVersionID,
						},
						{
							// Create commit
							"schemaVersionId": initialSchemaVersionID,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
