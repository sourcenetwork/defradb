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

func TestCollectionVersionOneOne_NoPrimary_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionOneOne_TwoPrimaries_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionOneOne_SelfUsingActualName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						boss: User @primary
						minion: User
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
								Name:         "_minionID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "boss",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
								IsPrimary:    true,
							},
							{
								Name:         "minion",
								Kind:         client.NewSelfKind("", false),
								RelationName: immutable.Some("user_user"),
							},
						},
						Indexes: []client.IndexDescription{
							{
								Name:   "User__bossID_ASC",
								ID:     1,
								Unique: true,
								Fields: []client.IndexedFieldDescription{
									{Name: "_bossID"},
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
								"name": "minion",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "User",
								},
							},
							Field{
								"name": "_minionID",
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
