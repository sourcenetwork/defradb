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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryWithP2PReplicatedDocAtOlderSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.SchemaPatch{
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
					SourceSchemaVersionID:      "bafkreiadnck34zzbwayjw3aeubw7eg4jmgtwoibu35tkxbjpar5rzxkdpu",
					DestinationSchemaVersionID: "bafkreibzqyjmyjs7vyo2q4h2tv5rbdbe4lv7tjbl5esilmobhgclia2juy",
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Node 0 should yield results as they were defined, as the newer schema version is
				// unknown to this node.
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
			testUtils.Request{
				// Node 1 should yield results migrated to the new schema version.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						// John has been migrated up to the newer schema version on node 1
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQueryWithP2PReplicatedDocAtNewerSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.SchemaPatch{
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
					SourceSchemaVersionID:      "bafkreiadnck34zzbwayjw3aeubw7eg4jmgtwoibu35tkxbjpar5rzxkdpu",
					DestinationSchemaVersionID: "bafkreibzqyjmyjs7vyo2q4h2tv5rbdbe4lv7tjbl5esilmobhgclia2juy",
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Node 0 should yield results as they were defined
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
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
				// Node 1 should yield results migrated down to the old schema version.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						// John has been migrated down to the older schema version on node 1
						// clearing the verified field
						"verified": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQueryWithP2PReplicatedDocAtMuchNewerSchemaVersionWithSchemaHistoryGap(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Patch node 0 only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.SchemaPatch{
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
				// There is no migration from version 1 to 2, thus node 1 has no knowledge of schema version 2.
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreih6o2jyurelxtpbg66gk23pio2tq6o3aed334z6w2u3qwve3at7ku",
					DestinationSchemaVersionID: "bafkreihv4ktjwzyhhkmas5iz4q7cawet4aeurqci33i66wr225l5pet4qu",
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Node 1 should also yield the synced doc, even though there was a gap in the schema version history
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
