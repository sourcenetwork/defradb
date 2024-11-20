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

// This test asserts that prefixes are being passed correctly through the new Lens fetcher.
func TestSchemaMigrationQueryByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, query by docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-48c8dacd-58ab-5fd5-8bbf-91bd823f4d5e
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
					SourceSchemaVersionID:      "bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe",
					DestinationSchemaVersionID: "bafkreiahhaeagyfsxaxmv3d665qvnbtyn3ts6jshhghy5bijwztbe7efpq",
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
					Users (docID: "bae-48c8dacd-58ab-5fd5-8bbf-91bd823f4d5e") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Shahzad",
							"verified": true,
						},
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
func TestSchemaMigrationQueryMultipleQueriesByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, multiple queries by docID",
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
				// bae-48c8dacd-58ab-5fd5-8bbf-91bd823f4d5e
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				// bae-3a7df128-bfa9-559a-a9c5-96f2bf6d1038
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.CreateDoc{
				// bae-5622129c-b893-5768-a3f4-8f745db4cc04
				Doc: `{
					"name": "Chris"
				}`,
			},
			testUtils.CreateDoc{
				// bae-6845cfdf-cb0f-56a3-be3a-b5a67be5fbdc
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				// bae-38a4ebb2-583a-5b6e-8e90-a6fe9e13be06
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.CreateDoc{
				// bae-4d2c0f6e-af73-54d9-ac8a-a419077ea1e5
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
					SourceSchemaVersionID:      "bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe",
					DestinationSchemaVersionID: "bafkreiahhaeagyfsxaxmv3d665qvnbtyn3ts6jshhghy5bijwztbe7efpq",
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
					Users (docID: "bae-48c8dacd-58ab-5fd5-8bbf-91bd823f4d5e") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Shahzad",
							"verified": true,
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-3a7df128-bfa9-559a-a9c5-96f2bf6d1038") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Fred",
							"verified": true,
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-5622129c-b893-5768-a3f4-8f745db4cc04") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Chris",
							"verified": true,
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-6845cfdf-cb0f-56a3-be3a-b5a67be5fbdc") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"verified": true,
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-38a4ebb2-583a-5b6e-8e90-a6fe9e13be06") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Islam",
							"verified": true,
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users (docID: "bae-4d2c0f6e-af73-54d9-ac8a-a419077ea1e5") {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "Dave",
							"verified": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
