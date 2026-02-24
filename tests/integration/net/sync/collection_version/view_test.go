// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSyncColVersion_WithView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				NodeID: immutable.Some(0),
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddLens{
				NodeID: immutable.Some(0),
				Lens: model.Lens{
					// This transform will copy the value from `name` into the `fullName` field,
					// like an overly-complicated alias
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.AddView{
				NodeID: immutable.Some(0),
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						fullName: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID1}}"},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				NodeID:        immutable.Some(1),
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "UserView",
						// Synced Views are always non-materialized when they first come in
						IsMaterialized: false,
						// Synced collections are inactive when they first come in
						IsActive: false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "fullName",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
						/* There is no good way to dynamically get the transform id at the moment, so unfortunately
						   we need to disable this assertion for now.  TestSyncColVersion_WithView_CanBeActivatedAndQueried
						   does prove that the transform is synced however.
							Query: immutable.Some(client.QuerySource{
								Query: request.Select{
									Field: request.Field{
										Name: "Users",
									},
									ChildSelect: request.ChildSelect{
										Fields: []request.Selection{
											&request.Field{
												Name: "name",
											},
										},
									},
								},
								Transform: immutable.Some("bafyreieqlg5ubzscocd6toe6dyxuadh2us342wnsyitmqxszexh7vajqga"),
							}),
						*/
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithView_CanBeActivatedAndQueried(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddLens{
				NodeID: immutable.Some(0),
				Lens: model.Lens{
					// This transform will copy the value from `name` into the `fullName` field,
					// like an overly-complicated alias
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.AddView{
				NodeID: immutable.Some(0),
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						fullName: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID1}}"},
			},
			testUtils.SetActiveCollectionVersion{
				NodeID:    immutable.Some(1),
				VersionID: "{{.CollectionVersionID1}}",
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
					UserView {
						fullName
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"fullName": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
