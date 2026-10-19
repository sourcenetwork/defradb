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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSyncColVersion_WithNonNillableBool(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						active: Boolean!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "active",
								Kind: client.FieldKind_BOOL,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableInt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						age: Int!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "age",
								Kind: client.FieldKind_INT,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						score: Float64!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "score",
								Kind: client.FieldKind_FLOAT64,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						weight: Float32!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "weight",
								Kind: client.FieldKind_FLOAT32,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableString(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						name: String!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						createdAt: DateTime!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "createdAt",
								Kind: client.FieldKind_DATETIME,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableBlob(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						avatar: Blob!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "avatar",
								Kind: client.FieldKind_BLOB,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithNonNillableJSON(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						metadata: JSON!
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "metadata",
								Kind: client.FieldKind_JSON,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
