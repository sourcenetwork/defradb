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
								Kind:         immutable.Some[client.FieldKind](client.ObjectArrayKind("Dog")),
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
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("User")),
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
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("User")),
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
								Kind:         immutable.Some[client.FieldKind](client.ObjectArrayKind("User")),
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
								Kind:         immutable.Some[client.FieldKind](client.ObjectArrayKind("User")),
								RelationName: immutable.Some("user_user"),
							},
							{
								Name:         "b",
								ID:           2,
								Kind:         immutable.Some[client.FieldKind](client.ObjectKind("User")),
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
