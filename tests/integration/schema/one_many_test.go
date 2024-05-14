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
	"github.com/sourcenetwork/defradb/client/request"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaOneMany_Primary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						dogs: [Dog]
					}
					type Dog {
						name: String
						owner: User @primary
					}
				`,
				ExpectedResults: []client.CollectionDescription{
					{
						Name: immutable.Some("User"),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "dogs",
								ID:           1,
								Kind:         immutable.Some[client.FieldKind](client.NewCollectionKind(2, true)),
								RelationName: immutable.Some("dog_user"),
							},
							{
								Name: "name",
								ID:   2,
							},
						},
					},
					{
						Name: immutable.Some("Dog"),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name: "name",
								ID:   1,
							},
							{
								Name:         "owner",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.NewCollectionKind(1, false)),
								RelationName: immutable.Some("dog_user"),
							},
							{
								Name:         "owner_id",
								ID:           3,
								Kind:         immutable.Some[client.FieldKind](client.ScalarKind(client.FieldKind_DocID)),
								RelationName: immutable.Some("dog_user"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneMany_SelfReferenceOneFieldLexographicallyFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						a: User
						b: [User]
					}
				`,
				ExpectedResults: []client.CollectionDescription{
					{
						Name: immutable.Some("User"),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "a",
								ID:           1,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", false)),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "a_id",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.ScalarKind(client.FieldKind_DocID)),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "b",
								ID:           3,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", true)),
								RelationName: immutable.Some("user_user"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneMany_SelfReferenceManyFieldLexographicallyFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						b: User
						a: [User]
					}
				`,
				ExpectedResults: []client.CollectionDescription{
					{
						Name: immutable.Some("User"),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "a",
								ID:           1,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", true)),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "b",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", false)),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "b_id",
								ID:           3,
								Kind:         immutable.Some[client.FieldKind](client.ScalarKind(client.FieldKind_DocID)),
								RelationName: immutable.Some("user_user"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneMany_SelfUsingActualName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				// Note: The @primary directive is required due to
				// https://github.com/sourcenetwork/defradb/issues/2620
				// it should be removed when that ticket is closed.
				Schema: `
					type User {
						boss: User @primary
						minions: [User]
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name: immutable.Some("User"),
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
							},
							{
								Name:         "boss",
								ID:           1,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", false)),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "boss_id",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "minions",
								ID:           3,
								Kind:         immutable.Some[client.FieldKind](client.NewSelfKind("", true)),
								RelationName: immutable.Some("user_user"),
							},
						},
					},
				},
			},
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "User",
						Root:      "bafkreiabowhib7ym6wumgywbc6f747e4u665ddxa5q43povrmjk4rmkzfi",
						VersionID: "bafkreiabowhib7ym6wumgywbc6f747e4u665ddxa5q43povrmjk4rmkzfi",
						Fields: []client.SchemaFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "boss",
								Kind: client.NewSelfKind("", false),
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "boss_id",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
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
						"fields": append(DefaultFields,
							Field{
								"name": "boss",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
							Field{
								"name": "boss_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
							Field{
								"name": "minions",
								"type": map[string]any{
									"kind": "LIST",
									"name": nil,
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
