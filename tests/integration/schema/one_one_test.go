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

func TestSchemaOneOne_NoPrimary_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						dog: Dog
					}
					type Dog {
						name: String
						owner: User
					}
				`,
				// This error is dependent upon the order in which definitions are validated, so
				// we only assert that the error is the correct type, and do not check the key-values
				ExpectedError: "relation missing field",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneOne_TwoPrimaries_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						dog: Dog @primary
					}
					type Dog {
						name: String
						owner: User @primary
					}
				`,
				ExpectedError: "relation can only have a single field set as primary",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneOne_SelfUsingActualName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						boss: User @primary
						minion: User
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
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("User")),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "boss_id",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "minion",
								ID:           3,
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("User")),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "minion_id",
								ID:           4,
								Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
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
								Kind: client.ObjectKind("User"),
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
								"name": "minion",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
							Field{
								"name": "minion_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
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
