// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// This test asserts that spans are being passed correctly through the new Lens fetcher.
func TestSchemaMigrationQueryByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, query by key",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-d7546ac1-c133-5853-b866-9b9f926fe7e5
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreig3zt63qt7bkji47etyu2sqtzroa3tcfdxgwqc3ka2ijy63refq3a",
					DestinationSchemaVersionID: "bafkreia4m6sn2rfypj2velvwpyude22fcb5jyfzum2eh3cdzg4a3myj5nu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-d7546ac1-c133-5853-b866-9b9f926fe7e5") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Shahzad",
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test asserts that lenses are being correctly returned to the pool for reuse after
// fetch completion. Querying by docID should mean that the fetcher only scans the docID
// prefix, and thus will only migrate a single document per query (unlike filters etc which
// will migrate all documents at the time of writing). If the return mechanic was very faulty
// then this test *should* deadlock.
//
// This behaviour should be covered more in-depth by unit tests, as it would be particularly
// bad if it broke and is fairly encumbersome to fully test via our current integration test
// framework.
//
// At the time of writing, the lens pool size is hardcoded to 5, so we should test with 6
// documents/queries, if the size changes so should this test.
func TestSchemaMigrationQueryMultipleQueriesByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, multiple queries by key",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			// We want 6 documents, and 6 queries, as lens pool is limited to 5
			// and we want to make sure that lenses are being correctly returned
			// to the pool for reuse after.
			testUtils.CreateDoc{
				// bae-d7546ac1-c133-5853-b866-9b9f926fe7e5
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				// bae-92393ad0-07b6-5753-8dbb-19c9c41374ed
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.CreateDoc{
				// bae-403d7337-f73e-5c81-8719-e853938c8985
				Doc: `{
					"name": "Chris"
				}`,
			},
			testUtils.CreateDoc{
				// bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				// bae-3f1174ba-d9bc-5a6a-b0bc-8f19581f199d
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.CreateDoc{
				// bae-0698bda7-2c69-5028-a26a-0a1c491b793b
				Doc: `{
					"name": "Dave"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreig3zt63qt7bkji47etyu2sqtzroa3tcfdxgwqc3ka2ijy63refq3a",
					DestinationSchemaVersionID: "bafkreia4m6sn2rfypj2velvwpyude22fcb5jyfzum2eh3cdzg4a3myj5nu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-d7546ac1-c133-5853-b866-9b9f926fe7e5") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Shahzad",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-92393ad0-07b6-5753-8dbb-19c9c41374ed") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Fred",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-403d7337-f73e-5c81-8719-e853938c8985") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Chris",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "John",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-3f1174ba-d9bc-5a6a-b0bc-8f19581f199d") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Islam",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-0698bda7-2c69-5028-a26a-0a1c491b793b") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "Dave",
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
