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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestCollectionMigrationQueryWithP2PReplicatedDocAtOlderSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.PatchCollection{
				// Patch node 1 only
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				// Register the migration on both nodes.
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreiabmrtgxy5dgotuc53gfaamuqhlzugyeetzbuv7s3x6ufmlr5ylga",
					DestinationCollectionVersionID: "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm",
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
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Node 0 should yield results as they were defined, as the newer collection version is
				// unknown to this node.
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			&action.Request{
				// Node 1 should yield results migrated to the new collection version.
				NodeID: immutable.Some(1),
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
							// John has been migrated up to the newer collection version on node 1
							"verified": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQueryWithP2PReplicatedDocAtMuchOlderSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.PatchCollection{
				// Patch node 1 only
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			&action.PatchCollection{
				// Patch node 1 only
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "address", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				// Register the migration on both nodes.
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreiabmrtgxy5dgotuc53gfaamuqhlzugyeetzbuv7s3x6ufmlr5ylga",
					DestinationCollectionVersionID: "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm",
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
			testUtils.ConfigureMigration{
				// Register the migration on both nodes.
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm",
					DestinationCollectionVersionID: "bafyreibdug5imopgzjyjclddzpkl3uxua4qz2fhc3myirsorg56hdrijyu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "name",
									"value": "Fred",
								},
							},
						},
					},
				},
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Node 0 should yield results as they were defined, as the newer collection version is
				// unknown to this node.
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			&action.Request{
				// Node 1 should yield results migrated to the new collection version.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							// John has been migrated up to the newer collection version on node 1
							"verified": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQueryWithP2PReplicatedDocAtNewerSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			&action.PatchCollection{
				// Patch node 0 only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				// Register the migration on both nodes.
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreiabmrtgxy5dgotuc53gfaamuqhlzugyeetzbuv7s3x6ufmlr5ylga",
					DestinationCollectionVersionID: "bafyreidwvvr7kp5rqt7dbgzw55vuueovkjz6b2mlvz3rq2pxf22fqenzdm",
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
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Node 0 should yield results as they were defined
				NodeID: immutable.Some(0),
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
			&action.Request{
				// Node 1 should yield results migrated down to the old collection version.
				NodeID: immutable.Some(1),
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
							// John has been migrated down to the older collection version on node 1
							// clearing the verified field
							"verified": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQueryWithP2PReplicatedDocAtMuchNewerSchemaVersionWithSchemaHistoryGap(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				// Patch node 0 only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				// Patch node 0 only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				// Register a migration from version 2 to version 3 on both nodes.
				// There is no migration from version 1 to 2, thus node 1 has no knowledge of collection version 2.
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "bafyreigqfjat435ghyt66tdaucp7oi2mke5jafx3jw3rozanopihr2vf44",
					DestinationCollectionVersionID: "bafyreiabghlustwur2y3pdxmoyq35rxcxg7bbgx6hxe2vezqow3q27g6za",
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
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Node 1 should also yield the synced doc, even though there was a gap in the collection version history
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
