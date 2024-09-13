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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaSelfReferenceSimple_SchemaHasSimpleSchemaID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "User",
						Root:      "bafkreifchjktkdtha7vkcqt6itzsw6lnzfyp7ufws4s32e7vigu7akn2q4",
						VersionID: "bafkreifchjktkdtha7vkcqt6itzsw6lnzfyp7ufws4s32e7vigu7akn2q4",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "boss",
								Typ:  client.LWW_REGISTER,
								// Simple self kinds do not contain a base ID, as there is only one possible value
								// that they could hold
								Kind: client.NewSelfKind("", false),
							},
							{
								Name: "boss_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
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
			testUtils.SchemaUpdate{
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
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name: "Dog",
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						Root:      "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-0",
						VersionID: "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-0",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "walker",
								Typ:  client.LWW_REGISTER,
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind: client.NewSelfKind("1", false),
							},
							{
								Name: "walker_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name: "User",
						// Note how Dog and User share the same base ID, but with a different index suffixed on
						// the end.
						Root:      "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-1",
						VersionID: "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-1",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								Typ:  client.LWW_REGISTER,
								// Because Dog and User form a circular dependency tree, the relation is declared
								// as a SelfKind, with the index identifier of User being held in the relation kind.
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "hosts_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
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
			testUtils.SchemaUpdate{
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
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name: "Cat",
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						Root:      "bafkreiacf7kjwlw32eiizyy6awdnfrnn7edaptp2chhfc5xktgxvrccqsa-0",
						VersionID: "bafkreiacf7kjwlw32eiizyy6awdnfrnn7edaptp2chhfc5xktgxvrccqsa-0",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "loves",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("1", false),
							},
							{
								Name: "loves_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "tolerates",
								Typ:  client.LWW_REGISTER,
								// This relationship reaches out of the Cat/Dog circle, and thus must be of type SchemaKind,
								// specified with the full User ID (including the `-1` index suffixed).
								Kind: client.NewSchemaKind("bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-1", false),
							},
							{
								Name: "tolerates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name: "Mouse",
						// Cat and Mouse share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Dog/User base ID.
						Root:      "bafkreiacf7kjwlw32eiizyy6awdnfrnn7edaptp2chhfc5xktgxvrccqsa-1",
						VersionID: "bafkreiacf7kjwlw32eiizyy6awdnfrnn7edaptp2chhfc5xktgxvrccqsa-1",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hates",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "hates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name: "Dog",
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						Root:      "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-0",
						VersionID: "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-0",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "walker",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("1", false),
							},
							{
								Name: "walker_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name: "User",
						// Dog and User share the same base ID, but with a different index suffixed on
						// the end.  This base must be different to the Cat/Mouse base ID.
						Root:      "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-1",
						VersionID: "bafkreichlth4ajgalengyv3hnmqnxa4vhnv5f34a3gzwh2jaajqb2yxd4i-1",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "hosts_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
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
			testUtils.SchemaUpdate{
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
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Cat",
						Root:      "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-0",
						VersionID: "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-0",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "loves",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("2", false),
							},
							{
								Name: "loves_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "tolerates",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("3", false),
							},
							{
								Name: "tolerates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Dog",
						Root:      "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-1",
						VersionID: "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-1",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "walker",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("3", false),
							},
							{
								Name: "walker_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Mouse",
						Root:      "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-2",
						VersionID: "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-2",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hates",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "hates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "User",
						Root:      "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-3",
						VersionID: "bafkreibykyk7nm7hbh44rnyqc6glt7d73dpnn3ttwmichwdqydiajjh3ea-3",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "feeds",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "feeds_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("1", false),
							},
							{
								Name: "hosts_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
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
			testUtils.SchemaUpdate{
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
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Cat",
						Root:      "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-0",
						VersionID: "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-0",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "loves",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("2", false),
							},
							{
								Name: "loves_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "tolerates",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("3", false),
							},
							{
								Name: "tolerates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Dog",
						Root:      "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-1",
						VersionID: "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-1",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "licks",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("2", false),
							},
							{
								Name: "licks_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "walker",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("3", false),
							},
							{
								Name: "walker_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Mouse",
						Root:      "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-2",
						VersionID: "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-2",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hates",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("0", false),
							},
							{
								Name: "hates_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "User",
						Root:      "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-3",
						VersionID: "bafkreidetmki4jtod5jfmromvcz2vd75j6t6g3vnw3aenlv7znludye4ru-3",
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Typ:  client.NONE_CRDT,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "hosts",
								Typ:  client.LWW_REGISTER,
								Kind: client.NewSelfKind("1", false),
							},
							{
								Name: "hosts_id",
								Typ:  client.LWW_REGISTER,
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
