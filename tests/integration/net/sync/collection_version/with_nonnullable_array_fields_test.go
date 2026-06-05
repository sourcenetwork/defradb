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

func TestSyncColVersion_WithNonNullableBoolArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						flags: [Boolean!]
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
								Name: "flags",
								Kind: client.FieldKind_BOOL_ARRAY,
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

func TestSyncColVersion_WithNonNullableIntArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						scores: [Int!]
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
								Name: "scores",
								Kind: client.FieldKind_INT_ARRAY,
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

func TestSyncColVersion_WithNonNullableFloat64Array(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						ratings: [Float64!]
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
								Name: "ratings",
								Kind: client.FieldKind_FLOAT64_ARRAY,
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

func TestSyncColVersion_WithNonNullableFloat32Array(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						weights: [Float32!]
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
								Name: "weights",
								Kind: client.FieldKind_FLOAT32_ARRAY,
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

func TestSyncColVersion_WithNonNullableStringArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						tags: [String!]
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
								Name: "tags",
								Kind: client.FieldKind_STRING_ARRAY,
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
