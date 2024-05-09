// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

func createAuthorBooksSchemaWithPolicyAndCreateDocs() []any {
	return []any{
		testUtils.AddPolicy{
			Identity:         acpUtils.Actor1Identity,
			Policy:           bookAuthorPolicy,
			ExpectedPolicyID: "68a4e64d5034b8a0565a90cd36483de0d61e0ea2450cf57c1fa8d27cbbf17c2c",
		},
		testUtils.SchemaUpdate{
			Schema: `
				type Author @policy(
					id: "68a4e64d5034b8a0565a90cd36483de0d61e0ea2450cf57c1fa8d27cbbf17c2c",
					resource: "author"
				) {
					name: String
					age: Int @index
					verified: Boolean
					published: [Book]
				}

				type Book @policy(
					id: "68a4e64d5034b8a0565a90cd36483de0d61e0ea2450cf57c1fa8d27cbbf17c2c",
					resource: "author"
				) {
					name: String
					rating: Float @index
					author: Author
				}`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
			Doc: `{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`,
		},
		testUtils.CreateDoc{
			Identity:     acpUtils.Actor1Identity,
			CollectionID: 0,
			// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
			Doc: `{
				"name": "Cornelia Funke",
				"age": 62,
				"verified": false
			}`,
		},
		testUtils.CreateDoc{
			CollectionID: 1,
			Doc: `{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`,
		},
		testUtils.CreateDoc{
			Identity:     acpUtils.Actor1Identity,
			CollectionID: 1,
			Doc: `{
				"name": "A Time for Mercy",
				"rating": 4.5,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`,
		},
		testUtils.CreateDoc{
			Identity:     acpUtils.Actor1Identity,
			CollectionID: 1,
			Doc: `{
				"name": "Theif Lord",
				"rating": 4.8,
				"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
			}`,
		},
	}
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (one-to-many) related doc without identity should not fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Request: `
					query {
						Author(filter: {
							published: {rating: {_gt: 3}}
						}) {
							name
							published {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"published": []map[string]any{
							{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithIdentity_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (one-to-many) related doc with identity should fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						Author(filter: {
							published: {rating: {_gt: 3}}
						}) {
							name
							published {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"published": []map[string]any{
							{
								"name": "Painted House",
							},
							{
								"name": "A Time for Mercy",
							},
						},
					},
					{
						"name": "Cornelia Funke",
						"published": []map[string]any{
							{
								"name": "Theif Lord",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithWrongIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (one-to-many) related doc with wrong identity should not fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						Author(filter: {
							published: {rating: {_gt: 3}}
						}) {
							name
							published {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"published": []map[string]any{
							{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (many-to-one) related doc without identity should not fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Request: `
					query {
						Book(filter: {
							author: {age: {_gt: 60}}
						}) {
							name
							author {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithIdentity_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (many-to-one) related doc with identity should fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						Book(filter: {
							author: {age: {_gt: 60}}
						}) {
							name
							author {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Theif Lord",
						"author": map[string]any{
							"name": "Cornelia Funke",
						},
					},
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
					{
						"name": "A Time for Mercy",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithWrongIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test ACP with index: upon querying private (many-to-one) related doc without identity should not fetch",
		Actions: []any{
			createAuthorBooksSchemaWithPolicyAndCreateDocs(),
			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						Book(filter: {
							author: {age: {_gt: 60}}
						}) {
							name
							author {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
