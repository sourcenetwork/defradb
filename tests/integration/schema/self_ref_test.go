// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaSelfReferenceSimple_SchemaHasSimpleSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						boss: User
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "User") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "User",
						"fields": DefaultFields.Append(
							Field{
								"name": "boss_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Append(
							Field{
								"name": "boss",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
						).Tidy(),
					},
				},
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						CollectionID:   "bafyreifwsspsoii73siptvgtugaz7maw3hyqsxghzw7m62waukq6bmzcmi",
						VersionID:      "bafyreifwsspsoii73siptvgtugaz7maw3hyqsxghzw7m62waukq6bmzcmi",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "boss",
								// Simple self kinds do not contain a base ID, as there is only one possible value
								// that they could hold
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "boss_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoTypes_SchemaHasComplexSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// The two primary relations form a circular two-collection self reference
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
						walks: Dog @relation(name:"walkies")
					}
					type Dog {
						host: User @relation(name:"hosts")
						walker: User @primary @relation(name:"walkies")
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "User") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "User",
						"fields": DefaultFields.Append(
							Field{
								"name": "hosts_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Append(
							Field{
								"name": "hosts",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "Dog",
								},
							},
						).Append(
							Field{
								"name": "walks_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Append(
							Field{
								"name": "walks",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "Dog",
								},
							},
						).Tidy(),
					},
				},
			},
			testUtils.IntrospectionRequest{
				Request: `
						query {
							__type (name: "Dog") {
								name
								fields {
									name
									type {
									name
									kind
									}
								}
							}
						}
					`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Dog",
						"fields": DefaultFields.Append(
							Field{
								"name": "host_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Append(
							Field{
								"name": "host",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
						).Append(
							Field{
								"name": "walker_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Append(
							Field{
								"name": "walker",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
						).Tidy(),
					},
				},
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						VersionID:      "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "walks_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						VersionID:      "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "host_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name: "walker",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "walker_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoTypes_SchemaHasComplexSchemaID_SingleSidedRelations(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				// The two primary relations form a circular two-collection self reference
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
					}
					type Dog {
						walker: User @primary @relation(name:"walkies")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						VersionID:      "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						VersionID:      "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "walker",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "walker_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoPairsOfTwoTypes_SchemasHaveDifferentComplexSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// - User and Dog form a circular dependency.
				// - Cat and Mouse form a another circular dependency.
				// - There is a relationship from Cat to User, this does not form a circular dependency
				// between the two (User/Dog and Cat/Mouse) circles, this is included to ensure that
				// the code does not incorrectly merge the User/Dog and Cat/Mouse circles into a single
				// circle.
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
						walks: Dog @relation(name:"walkies")
						toleratedBy: Cat @relation(name:"tolerates")
					}
					type Dog {
						host: User @relation(name:"hosts")
						walker: User @primary @relation(name:"walkies")
					}
					type Cat {
						loves: Mouse @primary @relation(name:"loves")
						hatedBy: Mouse @relation(name:"hates")
						tolerates: User @primary @relation(name:"tolerates")
					}
					type Mouse {
						lovedBy: Cat @relation(name:"loves")
						hates: Cat @primary @relation(name:"hates")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      1,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						VersionID:      "bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hosts",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "toleratedBy",
								Kind:         client.NewCollectionKind("bafyreienyzedhnqtmvtn2nv2e2dl2vquxyiwc2c45faflmlg27yg4amqc4", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "toleratedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "walks_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
						},
					},
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreicksowygm76pakx5vjlljodjwwltdzayrwj6qie554j3ygweqwqn4",
							RelativeID:      1,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						VersionID:      "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "hates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "lovedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreicksowygm76pakx5vjlljodjwwltdzayrwj6qie554j3ygweqwqn4",
							RelativeID:      0,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreienyzedhnqtmvtn2nv2e2dl2vquxyiwc2c45faflmlg27yg4amqc4",
						VersionID:      "bafyreienyzedhnqtmvtn2nv2e2dl2vquxyiwc2c45faflmlg27yg4amqc4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "hatedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "loves_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name: "tolerates",
								// This relationship reaches out of the Cat/Dog circle, and thus must be of type SchemaKind,
								// specified with the full User ID (including the `-1` index suffixed).
								Kind:         client.NewCollectionKind("bafyreibzwsd2nl3dq473lx3knf4g7yusnd5qktuxfg6kqcdnr3svrbjkb4", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreigqhogwrfyvqw33ujggwass2wgzbhqc2ttw2eny4doe42e2p4qyue",
							RelativeID:      0,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						VersionID:      "bafyreifb5zdzhynbrczhatx4tywarypl2v6mmo6cub74wimmxcf3xv7y24",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "host_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "walker_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoPairsOfTwoTypesJoinedByThirdCircle_SchemasAllHaveSameBaseSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// - User and Dog form a circular dependency.
				// - Cat and Mouse form a another circular dependency.
				// - User and Cat form a circular dependency - this circle overlaps with the two otherwise
				// independent User/Dog and Cat/Mouse circles, causing the 4 types to be locked together in
				// a larger circle (a relationship DAG cannot be formed) - all 4 types must thus share the
				// same base ID.
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
						walks: Dog @relation(name:"walkies")
						toleratedBy: Cat @relation(name:"tolerates")
						feeds: Cat @primary @relation(name:"feeds")
					}
					type Dog {
						host: User @relation(name:"hosts")
						walker: User @primary @relation(name:"walkies")
					}
					type Cat {
						loves: Mouse @primary @relation(name:"loves")
						hatedBy: Mouse @relation(name:"hates")
						tolerates: User @primary @relation(name:"tolerates")
						fedBy: User @relation(name:"feeds")
					}
					type Mouse {
						lovedBy: Cat @relation(name:"loves")
						hates: Cat @primary @relation(name:"hates")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih6p2qt3p3kcehh34uao5y5safkhskdfbshlaje63er66gryj65uq",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						VersionID:      "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "hates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "lovedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
						},
					},
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih6p2qt3p3kcehh34uao5y5safkhskdfbshlaje63er66gryj65uq",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreidrtmlxueujymjwidnysuqodrwr3pknivbefb3fuyxe7qv7q6evl4",
						VersionID:      "bafyreidrtmlxueujymjwidnysuqodrwr3pknivbefb3fuyxe7qv7q6evl4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "feeds",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("feeds"),
								IsPrimary:    true,
							},
							{
								Name:         "feeds_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("feeds"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "toleratedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "toleratedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "walks_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih6p2qt3p3kcehh34uao5y5safkhskdfbshlaje63er66gryj65uq",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreiducawb6xng7urrbpnf5ytjs6kbmmibwltkgrnljlaww7tm6imvju",
						VersionID:      "bafyreiducawb6xng7urrbpnf5ytjs6kbmmibwltkgrnljlaww7tm6imvju",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "host_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "walker_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih6p2qt3p3kcehh34uao5y5safkhskdfbshlaje63er66gryj65uq",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreig2bcrptryofxluqc4sfynbdao3jggtocgi3wmxav6ixb3xydqbxy",
						VersionID:      "bafyreig2bcrptryofxluqc4sfynbdao3jggtocgi3wmxav6ixb3xydqbxy",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "fedBy",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("feeds"),
							},
							{
								Name:         "fedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("feeds"),
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "hatedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "loves_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoPairsOfTwoTypesJoinedByThirdCircleAcrossAll_SchemasAllHaveSameBaseSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// - User and Dog form a circular dependency.
				// - Cat and Mouse form a another circular dependency.
				// - A larger circle is formed by bridging the two (User/Dog and Cat/Mouse) circles
				// at different points in the same direction - this circle forms from
				// User=>Dog=>Mouse=>Cat=>User=>etc.  This test ensures that the two independent circles do not
				// confuse the code into ignoring the larger circle.
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
						walks: Dog @relation(name:"walkies")
						toleratedBy: Cat @relation(name:"tolerates")
					}
					type Dog {
						host: User @relation(name:"hosts")
						walker: User @primary @relation(name:"walkies")
						licks: Mouse @primary @relation(name:"licks")
					}
					type Cat {
						loves: Mouse @primary @relation(name:"loves")
						hatedBy: Mouse @relation(name:"hates")
						tolerates: User @primary @relation(name:"tolerates")
					}
					type Mouse {
						lovedBy: Cat @relation(name:"loves")
						hates: Cat @primary @relation(name:"hates")
						lickedBy: Dog @relation(name:"licks")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidcymuojy4qpyjuzbjdjekn3jea4fu776zflycowgnbqo3wp2jrom",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						VersionID:      "bafyreic5bld2kagpt7o7qc2olgxoovppry3xn4tbjnnwoeci532hax5vji",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "hates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lickedBy",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("licks"),
							},
							{
								Name:         "lickedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("licks"),
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "lovedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidcymuojy4qpyjuzbjdjekn3jea4fu776zflycowgnbqo3wp2jrom",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreicf444olomeq4xlbuxqg2377r653etyjkshlnadcyyaegf3exy3bm",
						VersionID:      "bafyreicf444olomeq4xlbuxqg2377r653etyjkshlnadcyyaegf3exy3bm",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "host_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "licks",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("licks"),
								IsPrimary:    true,
							},
							{
								Name:         "licks_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("licks"),
								IsPrimary:    true,
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "walker_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
					},
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidcymuojy4qpyjuzbjdjekn3jea4fu776zflycowgnbqo3wp2jrom",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreif3dwkklu53xczlcs5okaexpn2fsnhqpqurg2vatwwwn5qakojwnm",
						VersionID:      "bafyreif3dwkklu53xczlcs5okaexpn2fsnhqpqurg2vatwwwn5qakojwnm",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hosts",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "hosts_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "toleratedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "toleratedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "walks_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidcymuojy4qpyjuzbjdjekn3jea4fu776zflycowgnbqo3wp2jrom",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreig2bcrptryofxluqc4sfynbdao3jggtocgi3wmxav6ixb3xydqbxy",
						VersionID:      "bafyreig2bcrptryofxluqc4sfynbdao3jggtocgi3wmxav6ixb3xydqbxy",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "hatedBy_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "loves_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates_id",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
