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
						CollectionID:   "bafyreicuxpdrri4wwdknhbchhdii6tu4myqlhspv3s2c3pci7jt7qc3zua",
						VersionID:      "bafyreicuxpdrri4wwdknhbchhdii6tu4myqlhspv3s2c3pci7jt7qc3zua",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_bossID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name: "boss",
								// Simple self kinds do not contain a base ID, as there is only one possible value
								// that they could hold
								Kind:         client.NewSelfKind("", false),
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
								"name": "_bossID",
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
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						VersionID:      "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "_walksID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
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
								Name:         "walks",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("walkies"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "User__hostsID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hostsID"},
								},
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						VersionID:      "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "_walkerID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("1", false),
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
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Dog__walkerID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_walkerID"},
								},
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
								"name": "_hostsID",
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
								"name": "_walksID",
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
								"name": "_hostID",
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
								"name": "_walkerID",
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
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      1,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						VersionID:      "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name: "hosts",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      0,
						}),
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						CollectionID:   "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						VersionID:      "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_walkerID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name: "walker",
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind:         client.NewSelfKind("1", false),
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
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      1,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						VersionID:      "bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "_walksID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "hosts",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "toleratedBy",
								Kind:         client.NewCollectionKind("bafyreiglwckweqrmegu2nzhsyosjn3cyt6r2tr6g6x2jw2s33t5zbuz3my", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("walkies"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "User__hostsID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hostsID"},
								},
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreihwm4xrfdaa44cdu7jqowq56ct2vbhkomd4gm3rc7atxozzjeaiwq",
							RelativeID:      0,
						}),
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						CollectionID:   "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						VersionID:      "bafyreieatejtfovgcmoec46cvgy3qmlk3br764buuzauildkpmfqanby4u",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "_walkerID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Dog__walkerID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_walkerID"},
								},
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreig27lu3bkeqsidbml3xfmhb5za5dorucotdgbammezhr2kwkrosxa",
							RelativeID:      0,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreiglwckweqrmegu2nzhsyosjn3cyt6r2tr6g6x2jw2s33t5zbuz3my",
						VersionID:      "bafyreiglwckweqrmegu2nzhsyosjn3cyt6r2tr6g6x2jw2s33t5zbuz3my",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hatedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "_lovesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name: "tolerates",
								// This relationship reaches out of the Cat/Dog circle, and thus must be of type SchemaKind,
								// specified with the full User ID (including the `-1` index suffixed).
								Kind:         client.NewCollectionKind("bafyreidpz32533bbfns2zv6v3c2c6hty2t5edf3vl2az4cndzrjwbn6fba", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Cat__lovesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_lovesID"},
								},
							},
							{
								Name:   "Cat__toleratesID_ASC",
								ID:     2,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_toleratesID"},
								},
							},
						},
					},
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreig27lu3bkeqsidbml3xfmhb5za5dorucotdgbammezhr2kwkrosxa",
							RelativeID:      1,
						}),
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						CollectionID:   "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						VersionID:      "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hatesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "_lovedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Mouse__hatesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hatesID"},
								},
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
							CollectionSetID: "bafyreibzcj6impd4v3sxn3fnth6t2dq26icretu5hd2dw73tpkudtecwg4",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreiegxkjktusmxjzreumct6h2zxrt5x7tjhpyyqcjewdhqnzx6hwob4",
						VersionID:      "bafyreiegxkjktusmxjzreumct6h2zxrt5x7tjhpyyqcjewdhqnzx6hwob4",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_feedsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("feeds"),
								IsPrimary:    true,
							},
							{
								Name:         "_hostsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "_walksID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "feeds",
								Kind:         client.NewSelfKind("0", false),
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
								Name:         "toleratedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "User__feedsID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_feedsID"},
								},
							},
							{
								Name:   "User__hostsID_ASC",
								ID:     2,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hostsID"},
								},
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreibzcj6impd4v3sxn3fnth6t2dq26icretu5hd2dw73tpkudtecwg4",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreidhvxz4t34lfb7utkshjczdewpoiwps77mzbllxawsgjvg7mhxl34",
						VersionID:      "bafyreidhvxz4t34lfb7utkshjczdewpoiwps77mzbllxawsgjvg7mhxl34",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "_walkerID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Dog__walkerID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_walkerID"},
								},
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreibzcj6impd4v3sxn3fnth6t2dq26icretu5hd2dw73tpkudtecwg4",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreihzlmak2gqmr72rcgkzi4emge45lnycs3oa4n35jzkfimfhlepaxu",
						VersionID:      "bafyreihzlmak2gqmr72rcgkzi4emge45lnycs3oa4n35jzkfimfhlepaxu",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_fedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("feeds"),
							},
							{
								Name:         "_hatedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "_lovesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "fedBy",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("feeds"),
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Cat__lovesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_lovesID"},
								},
							},
							{
								Name:   "Cat__toleratesID_ASC",
								ID:     2,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_toleratesID"},
								},
							},
						},
					},
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreibzcj6impd4v3sxn3fnth6t2dq26icretu5hd2dw73tpkudtecwg4",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						VersionID:      "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hatesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "_lovedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Mouse__hatesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hatesID"},
								},
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
							CollectionSetID: "bafyreieajbvoqehxtwivpnpnnig3tle5iv3itasgl4mtndxyveauxurgtq",
							RelativeID:      3,
						}),
						CollectionID:   "bafyreif5zqce5k2kl5wx23fjdtywxcdc5h2nauqwhy362smqrbk7xrrrje",
						VersionID:      "bafyreif5zqce5k2kl5wx23fjdtywxcdc5h2nauqwhy362smqrbk7xrrrje",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostsID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "_walksID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
							},
							{
								Name:         "hosts",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("hosts"),
								IsPrimary:    true,
							},
							{
								Name:         "toleratedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("tolerates"),
							},
							{
								Name:         "walks",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("walkies"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "User__hostsID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hostsID"},
								},
							},
						},
					},
					{
						Name: "Dog",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreieajbvoqehxtwivpnpnnig3tle5iv3itasgl4mtndxyveauxurgtq",
							RelativeID:      1,
						}),
						CollectionID:   "bafyreifujlparjdxteff2eftvarslel53sp7s3alit3z5sueufeewgjb3m",
						VersionID:      "bafyreifujlparjdxteff2eftvarslel53sp7s3alit3z5sueufeewgjb3m",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hostID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "_licksID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("licks"),
								IsPrimary:    true,
							},
							{
								Name:         "_walkerID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
							{
								Name:         "host",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("hosts"),
							},
							{
								Name:         "licks",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("licks"),
								IsPrimary:    true,
							},
							{
								Name:         "walker",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("walkies"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Dog__licksID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_licksID"},
								},
							},
							{
								Name:   "Dog__walkerID_ASC",
								ID:     2,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_walkerID"},
								},
							},
						},
					},
					{
						Name: "Cat",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreieajbvoqehxtwivpnpnnig3tle5iv3itasgl4mtndxyveauxurgtq",
							RelativeID:      0,
						}),
						CollectionID:   "bafyreihzlmak2gqmr72rcgkzi4emge45lnycs3oa4n35jzkfimfhlepaxu",
						VersionID:      "bafyreihzlmak2gqmr72rcgkzi4emge45lnycs3oa4n35jzkfimfhlepaxu",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hatedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "_lovesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "_toleratesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
							{
								Name:         "hatedBy",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("hates"),
							},
							{
								Name:         "loves",
								Kind:         client.NewSelfKind("2", false),
								RelationName: immutable.Some("loves"),
								IsPrimary:    true,
							},
							{
								Name:         "tolerates",
								Kind:         client.NewSelfKind("3", false),
								RelationName: immutable.Some("tolerates"),
								IsPrimary:    true,
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Cat__lovesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_lovesID"},
								},
							},
							{
								Name:   "Cat__toleratesID_ASC",
								ID:     2,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_toleratesID"},
								},
							},
						},
					},
					{
						Name: "Mouse",
						CollectionSet: immutable.Some(client.CollectionSetDescription{
							CollectionSetID: "bafyreieajbvoqehxtwivpnpnnig3tle5iv3itasgl4mtndxyveauxurgtq",
							RelativeID:      2,
						}),
						CollectionID:   "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						VersionID:      "bafyreigy2kzdolq2gq6dfm4s525gh4asp4mqexjonmyfdt6qbwlbc6v2vy",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "_hatesID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "_lickedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("licks"),
							},
							{
								Name:         "_lovedByID",
								Typ:          client.LWW_REGISTER,
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("loves"),
							},
							{
								Name:         "hates",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("hates"),
								IsPrimary:    true,
							},
							{
								Name:         "lickedBy",
								Kind:         client.NewSelfKind("1", false),
								RelationName: immutable.Some("licks"),
							},
							{
								Name:         "lovedBy",
								Kind:         client.NewSelfKind("0", false),
								RelationName: immutable.Some("loves"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "Mouse__hatesID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_hatesID"},
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
