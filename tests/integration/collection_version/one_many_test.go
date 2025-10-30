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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaOneMany_Primary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "dogs",
								Kind:         client.NewCollectionKind("bafyreih2kr3b6xijzkyv7yvjsg32selni5qehaejehf5vig7hpynjnbl5q", true),
								RelationName: immutable.Some("dog_user"),
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					{
						Name:           "Dog",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name:         "owner",
								Kind:         client.NewCollectionKind("bafyreibhpgygzsmki22sql5ejzcojrrxbc5iuhpydhdzxul5w2znc7zrgu", false),
								RelationName: immutable.Some("dog_user"),
								IsPrimary:    true,
							},
							{
								Name:         "owner_id",
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("dog_user"),
								Typ:          client.LWW_REGISTER,
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

func TestSchemaOneMany_SelfReferenceOneFieldLexographicallyFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						a: User
						b: [User]
					}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name:    "_docID",
								FieldID: "bafyreihqzhiz3iwro4jozp6kphq4sosg6ccoqcbiaf7rg5dmvea7aux55a",
								Kind:    client.FieldKind_DocID,
							},
							{
								Name:         "a",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "a_id",
								FieldID:      "bafyreieroxmzvqikc6mclepkm5tunroq6yuamr76f2tzu4nkwfy5au6lvi",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "b",
								Kind:         client.NewSelfKind("", true),
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
			&action.AddSchema{
				Schema: `
					type User {
						b: User
						a: [User]
					}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "a",
								Kind:         client.NewSelfKind("", true),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "b",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "b_id",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
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

func TestSchemaOneMany_SelfUsingActualName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				// Note: The @primary directive is required due to
				// https://github.com/sourcenetwork/defradb/issues/2620
				// it should be removed when that ticket is closed.
				Schema: `
					type User {
						boss: User @primary
						minions: [User]
					}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "boss",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "boss_id",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "minions",
								Kind:         client.NewSelfKind("", true),
								RelationName: immutable.Some("user_user"),
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
