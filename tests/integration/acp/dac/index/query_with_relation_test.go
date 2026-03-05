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

package test_acp_dac_index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func addAuthorBooksSchemaWithPolicyAndAddDocs() []any {
	return []any{
		testUtils.AddDACPolicy{
			Identity: testUtils.ClientIdentity(1),
			Policy:   bookAuthorPolicy,
		},
		&action.AddCollection{
			SDL: `
				type Author @policy(
					id: "{{.Policy0}}",
					resource: "author"
				) {
					name: String
					age: Int @index
					verified: Boolean
					published: [Book]
				}

				type Book @policy(
					id: "{{.Policy0}}",
					resource: "author"
				) {
					name: String
					rating: Float @index
					author: Author
				}`,
		},
		&action.AddDoc{
			CollectionID: 0,
			// bae-9e70648f-c722-5875-97f5-574ec6f703e9
			Doc: `{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`,
		},
		&action.AddDoc{
			Identity:     testUtils.ClientIdentity(1),
			CollectionID: 0,
			// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
			Doc: `{
				"name": "Cornelia Funke",
				"age": 62,
				"verified": false
			}`,
		},
		&action.AddDoc{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":      "Painted House",
				"rating":    4.9,
				"_authorID": testUtils.NewDocIndex(0, 0),
			},
		},
		&action.AddDoc{
			Identity:     testUtils.ClientIdentity(1),
			CollectionID: 1,
			DocMap: map[string]any{
				"name":      "A Time for Mercy",
				"rating":    4.5,
				"_authorID": testUtils.NewDocIndex(0, 0),
			},
		},
		&action.AddDoc{
			Identity:     testUtils.ClientIdentity(1),
			CollectionID: 1,
			DocMap: map[string]any{
				"name":      "Theif Lord",
				"rating":    4.8,
				"_authorID": testUtils.NewDocIndex(0, 1),
			},
		},
	}
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
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
				Results: map[string]any{
					"Author": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithIdentity_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
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
				Results: map[string]any{
					"Author": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateOneToManyRelatedDocWithWrongIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
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
				Results: map[string]any{
					"Author": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
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
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
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
				Results: map[string]any{
					"Book": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateManyToOneRelatedDocWithWrongIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addAuthorBooksSchemaWithPolicyAndAddDocs(),
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
