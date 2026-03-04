// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimple_WithFragments_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
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
