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
)

func TestQuerySimple_WithAliasEqualsFilterBlock_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: {UserAge: {_eq: 21}}}) {
						Name
						UserAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "John",
							"UserAge": int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithEmptyAlias_ShouldNotFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: {}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
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

func TestQuerySimple_WithNullAlias_ShouldFilterAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: null}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNonObjectAlias_ShouldFilterAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: 1}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNonExistantAlias_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: {UserAge: {_eq: 21}}}) {
						Name
						Age
					}
				}`,
				ExpectedError: `field or alias not found. Name: UserAge`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNonAliasedField_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_alias: {Age: {_eq: 32}}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithCompoundAlias_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {
						_and: [
							{_alias: {userAge: {_gt: 30}}},
							{_alias: {userAge: {_lt: 40}}}
						]
					}) {
						Name
						userAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Bob",
							"userAge": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasWithCompound_ShouldMatchFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {
						_alias: {
							_and: [
								{userAge: {_gt: 30}},
								{userAge: {_lt: 40}}
							]
						}
					}) {
						Name
						userAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Bob",
							"userAge": int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
