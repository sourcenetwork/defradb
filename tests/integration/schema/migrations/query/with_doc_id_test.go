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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// This test asserts that prefixes are being passed correctly through the new Lens fetcher.
func TestSchemaMigrationQueryByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, query by docID",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-8c6381d0-a558-5bac-8d60-67f78ba5ffb8
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
					DestinationSchemaVersionID: "bafyreig2nfxuzl3cob7txuvybcct6mmsylt57oirzsrehffkho6bdxlvwy",
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
					Users (docID: "bae-8c6381d0-a558-5bac-8d60-67f78ba5ffb8") {
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
			&action.AddSchema{
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
				// bae-8c6381d0-a558-5bac-8d60-67f78ba5ffb8
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				// bae-ce0ba9f3-4b91-5ae8-b5b2-bd667b4c443e
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.CreateDoc{
				// bae-01a26764-2c18-5049-b95e-70451f050459
				Doc: `{
					"name": "Chris"
				}`,
			},
			testUtils.CreateDoc{
				// bae-0623ed7c-0861-5995-a5d7-cce53642a83e
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				// bae-225b161c-25d7-5e2b-9014-ac3fc69fca5e
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.CreateDoc{
				// bae-bd4a814b-8338-5552-baec-73eb3d2a51bf
				Doc: `{
					"name": "Dave"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
					DestinationSchemaVersionID: "bafyreig2nfxuzl3cob7txuvybcct6mmsylt57oirzsrehffkho6bdxlvwy",
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
					Users (docID: "bae-8c6381d0-a558-5bac-8d60-67f78ba5ffb8") {
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
					Users (docID: "bae-ce0ba9f3-4b91-5ae8-b5b2-bd667b4c443e") {
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
					Users (docID: "bae-01a26764-2c18-5049-b95e-70451f050459") {
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
					Users (docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e") {
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
					Users (docID: "bae-225b161c-25d7-5e2b-9014-ac3fc69fca5e") {
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
					Users (docID: "bae-bd4a814b-8338-5552-baec-73eb3d2a51bf") {
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
