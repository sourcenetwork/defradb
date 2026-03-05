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

func TestSyncColVersion_WithPatchVersionOfUnknownCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				// Create Users on node 0 only, node 1 has no knowledge of it
				NodeID: immutable.Some(0),
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "age", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"bafyreics7adsddesun4kqqotr6g6c6ld2t7djlwcbrm4ftbhru3ayindy4"},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						// Synced collections are inactive when they first come in
						IsActive: false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					{
						Name:           "Users",
						IsMaterialized: true,
						// Synced collections are inactive when they first come in
						IsActive: false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
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

func TestSyncColVersion_WithPatchVersionOfKnownCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				// Create Users on both nodes, as the active version
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "age", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"bafyreics7adsddesun4kqqotr6g6c6ld2t7djlwcbrm4ftbhru3ayindy4"},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						// The original version was created directly on this node and was active,
						// receiving the new version has not changed this.
						IsActive: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					{
						Name:           "Users",
						IsMaterialized: true,
						// Synced collections are inactive when they first come in
						IsActive: false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "age",
								Kind: client.FieldKind_NILLABLE_INT,
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
