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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionOneMany_Primary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								Kind:         client.NewCollectionKind("bafyreicdzapaezl3xg22wjg5mu3in7wooqpn52jzazrigsayuan7eqlccq", true),
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
								Name:         "_ownerID",
								Kind:         client.FieldKind_DocID,
								RelationName: immutable.Some("dog_user"),
								Typ:          client.LWW_REGISTER,
								IsPrimary:    true,
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
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionOneMany_SelfReferenceOneFieldLexographicallyFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								Name:         "_aID",
								FieldID:      "bafyreibvvrhhuptmxjfzdk4glmw7aj2w6kv4cybrvg25ynrg6evas4j3ha",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "a",
								Kind:         client.NewSelfKind("", false),
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

func TestCollectionVersionOneMany_SelfReferenceManyFieldLexographicallyFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								Name:         "_bID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
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
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionOneMany_SelfUsingActualName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				// Note: The @primary directive is required due to
				// https://github.com/sourcenetwork/defradb/issues/2620
				// it should be removed when that ticket is closed.
				SDL: `
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
								Name:         "_bossID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "boss",
								Kind:         client.NewSelfKind("", false),
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
								"name": "_bossID",
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
