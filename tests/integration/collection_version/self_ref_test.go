// Copyright 2024 Democratized Data Foundation
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						CollectionID:   "bafyreia3se3uhyxtbazfheuxx7qtbgnd7lgiu4gsw5u33vyhbrenqvbcvm",
						VersionID:      "bafyreia3se3uhyxtbazfheuxx7qtbgnd7lgiu4gsw5u33vyhbrenqvbcvm",
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
						VersionID:      "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
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
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
						VersionID:      "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSelfReferenceTwoTypes_SchemaHasComplexSchemaID_SingleSidedRelations(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// The two primary relations form a circular two-collection self reference
				Schema: `
					type User {
						hosts: Dog @primary @relation(name:"hosts")
					}
					type Dog {
						walker: User @primary @relation(name:"walkies")
					}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
						VersionID:      "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
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
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
						VersionID:      "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      1,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
						VersionID:      "bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4",
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
								Kind:         client.NewCollectionKind("bafyreidwzjgomb4yvwmcaf2isbty4pjt3wu5heelxszwzp5msqfluxcosm", false),
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
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreih526pn4bdkru5u6lqrb457drmsmhxv6w74cblwy5minou5e5xanq",
							RelativeID:      0,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
						VersionID:      "bafyreieowx33bl3ohstvi47fzbapjfelznes7etf4gufktwcjfdcybwasi",
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
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreicz7y35n5vshwq33atlstpiriqlwqlcs6vou2s6izwmlmgjbu26zi",
							RelativeID:      0,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreidwzjgomb4yvwmcaf2isbty4pjt3wu5heelxszwzp5msqfluxcosm",
						VersionID:      "bafyreidwzjgomb4yvwmcaf2isbty4pjt3wu5heelxszwzp5msqfluxcosm",
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
								Kind:         client.NewCollectionKind("bafyreidtucrm7u3cyj5qyrlvimyet3mptezoycw6fqeuy6rg7nsi5u2un4", false),
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
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreicz7y35n5vshwq33atlstpiriqlwqlcs6vou2s6izwmlmgjbu26zi",
							RelativeID:      1,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
						VersionID:      "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreia3etvw4bglyoyazyifh3bcbcanx5xvztgmpsvx4dprxnrg27syn4",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreiccbgjelan5uoxa4lwob2lxwu672uuivqgo3k7t555hsrqp7etwxm",
						VersionID:      "bafyreiccbgjelan5uoxa4lwob2lxwu672uuivqgo3k7t555hsrqp7etwxm",
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
							CollectionSetID: "bafyreia3etvw4bglyoyazyifh3bcbcanx5xvztgmpsvx4dprxnrg27syn4",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreibwspwenuo2a7fynkbpyq5adxqnj5aprvestbzqpv3pqvqgmjgwdy",
						VersionID:      "bafyreibwspwenuo2a7fynkbpyq5adxqnj5aprvestbzqpv3pqvqgmjgwdy",
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
							CollectionSetID: "bafyreia3etvw4bglyoyazyifh3bcbcanx5xvztgmpsvx4dprxnrg27syn4",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreig3evelvkwlz7m3skb3vo3chiwf3navwnz4rcmuhxevq3j2hkjd6q",
						VersionID:      "bafyreig3evelvkwlz7m3skb3vo3chiwf3navwnz4rcmuhxevq3j2hkjd6q",
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
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreia3etvw4bglyoyazyifh3bcbcanx5xvztgmpsvx4dprxnrg27syn4",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
						VersionID:      "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name: "User",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidqk2lsb2v2ky2rz2k2tmqqs7btrhlktoscbsaez34qzo6azldk64",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreiajmn43lbrzpmyearh3vgvi3c2p7hy3dutbyvo5x67syno74cgesi",
						VersionID:      "bafyreiajmn43lbrzpmyearh3vgvi3c2p7hy3dutbyvo5x67syno74cgesi",
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
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidqk2lsb2v2ky2rz2k2tmqqs7btrhlktoscbsaez34qzo6azldk64",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreiehb4ykaynslqm47dsqwpixzpho3rxa3yopidgywqwz3u237q4kvy",
						VersionID:      "bafyreiehb4ykaynslqm47dsqwpixzpho3rxa3yopidgywqwz3u237q4kvy",
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
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidqk2lsb2v2ky2rz2k2tmqqs7btrhlktoscbsaez34qzo6azldk64",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreig3evelvkwlz7m3skb3vo3chiwf3navwnz4rcmuhxevq3j2hkjd6q",
						VersionID:      "bafyreig3evelvkwlz7m3skb3vo3chiwf3navwnz4rcmuhxevq3j2hkjd6q",
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
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreidqk2lsb2v2ky2rz2k2tmqqs7btrhlktoscbsaez34qzo6azldk64",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
						VersionID:      "bafyreibsop5phevv44lproydoa2l3cdfw5j4ru74ivgvm4tq4cojerv4xa",
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
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
