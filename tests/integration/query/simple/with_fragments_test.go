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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimple_WithFragments_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Named fragment spread returns the same fields as selecting them inline.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					firstUser: Users(limit: 1) {
						...UserInfo
					}
					lastUser: Users(limit: 1, offset: 1) {
						...UserInfo
					}
				}
				fragment UserInfo on Users {
  					Name
  					Age
				}`,
				Results: map[string]any{
					"firstUser": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(21),
						},
					},
					"lastUser": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNestedFragments_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Nested fragment spreads compose correctly and return all referenced fields.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						...UserWithNameAndAge
					}
				}
				fragment UserWithName on Users {
  					Name
				}
				fragment UserWithNameAndAge on Users {
  					...UserWithName
					Age
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(40),
						},
						{
							"Name": "Bob",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFragmentSpreadAndSelect_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Combining fragment spread with explicit field selection returns all fields without duplication.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						...UserAge
					}
				}
				fragment UserAge on Users {
					Age
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(40),
						},
						{
							"Name": "Bob",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMissingFragment_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Using an undefined fragment name returns a GraphQL validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						...UserInfo
					}
				}`,
				ExpectedError: `Unknown fragment "UserInfo".`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFragmentWithInvalidField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "A fragment referencing a non-existent field returns a schema validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						...UserInvalid
					}
				}
				fragment UserInvalid on Users {
					Score
				}`,
				ExpectedError: `Cannot query field "Score" on type "Users".`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFragmentWithAggregate_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "A fragment containing an aggregate field correctly returns the aggregated value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					...UserCount
				}
				fragment UserCount on Query {
					COUNT(Users: {})
				}`,
				Results: map[string]any{
					"COUNT": int64(2),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithFragmentWithVariables_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "A fragment used with GraphQL variables substitutes values correctly.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"filter": map[string]any{
						"Age": map[string]any{
							"_gt": int64(30),
						},
					},
				}),
				Request: `query($filter: UsersFilterArg!) {
					...UserFilter
				}
				fragment UserFilter on Query {
					Users(filter: $filter) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithInlineFragment_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "An inline fragment on the correct type returns the selected fields.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						... on Users {
							Name
							Age
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(40),
						},
						{
							"Name": "Bob",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
