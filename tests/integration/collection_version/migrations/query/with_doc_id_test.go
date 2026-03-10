// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
func TestCollectionMigrationQueryByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				// bae-d1536ab3-c3d8-5c3d-9622-087ee707fd99
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
					DestinationCollectionVersionID: "bafyreigqfjat435ghyt66tdaucp7oi2mke5jafx3jw3rozanopihr2vf44",
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-d1536ab3-c3d8-5c3d-9622-087ee707fd99") {
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
func TestCollectionMigrationQueryMultipleQueriesByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// We want 6 documents, and 6 queries, as lens pool is limited to 5
			// and we want to make sure that lenses are being correctly returned
			// to the pool for reuse after.
			&action.AddDoc{
				// bae-d1536ab3-c3d8-5c3d-9622-087ee707fd99
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.AddDoc{
				// bae-235c64e3-abf7-549c-9aff-971c8afdfa3f
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.AddDoc{
				// bae-eadc6f5f-a52b-57de-ad6c-e76315fff6bd
				Doc: `{
					"name": "Chris"
				}`,
			},
			&action.AddDoc{
				// bae-9b4d35b6-00f0-50df-8627-44cea1dbcf11
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				// bae-aa68c022-519a-50cf-8a91-2ff6d4349c90
				Doc: `{
					"name": "Islam"
				}`,
			},
			&action.AddDoc{
				// bae-81418211-7e0c-5e0c-8505-6288318c7248
				Doc: `{
					"name": "Dave"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
					DestinationCollectionVersionID: "bafyreigqfjat435ghyt66tdaucp7oi2mke5jafx3jw3rozanopihr2vf44",
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-d1536ab3-c3d8-5c3d-9622-087ee707fd99") {
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-235c64e3-abf7-549c-9aff-971c8afdfa3f") {
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-eadc6f5f-a52b-57de-ad6c-e76315fff6bd") {
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-9b4d35b6-00f0-50df-8627-44cea1dbcf11") {
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-aa68c022-519a-50cf-8a91-2ff6d4349c90") {
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
			&action.Request{
				Request: `query {
					Users (docID: "bae-81418211-7e0c-5e0c-8505-6288318c7248") {
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
