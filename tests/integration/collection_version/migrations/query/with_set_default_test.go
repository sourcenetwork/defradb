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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestCollectionMigrationQuery_WithSetDefaultToLatest_AppliesForwardMigration(t *testing.T) {
	collectionVersionID2 := "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "verified",
								"value": true,
							},
						},
					},
				}),
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: collectionVersionID2,
			},
			&action.Request{
				Request: `query {
					Users {
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQuery_WithSetDefaultToOriginal_AppliesInverseMigration(t *testing.T) {
	collectionVersionID1 := "bafyreiabmrtgxy5dgotuc53gfaamuqhlzugyeetzbuv7s3x6ufmlr5ylga"
	collectionVersionID2 := "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: collectionVersionID2,
			},
			// Create John using the new collection version
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      collectionVersionID1,
					DestinationCollectionVersionID: collectionVersionID2,
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
			// Set the collection version back to the original
			testUtils.SetActiveCollectionVersion{
				VersionID: collectionVersionID1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// The inverse lens migration has been applied, clearing the verified field
							"verified": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQuery_WithSetDefaultToOriginalVersionThatDocWasAddedAt_ClearsMigrations(t *testing.T) {
	collectionVersionID1 := "bafyreiabmrtgxy5dgotuc53gfaamuqhlzugyeetzbuv7s3x6ufmlr5ylga"
	collectionVersionID2 := "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm"

	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			// Create John using the original collection version
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"verified": false
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      collectionVersionID1,
					DestinationCollectionVersionID: collectionVersionID2,
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
			// Set the collection version back to the original
			testUtils.SetActiveCollectionVersion{
				VersionID: collectionVersionID1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// The inverse lens migration has not been applied, the document is returned as it was defined
							"verified": false,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
