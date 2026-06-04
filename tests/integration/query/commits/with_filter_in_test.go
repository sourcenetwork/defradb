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

package commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits_WithFilterFieldNameIn_ReturnsMatchingCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_in: ["age", "name"]}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "age"},
						{"fieldName": "name"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameInComposite_ReturnsCompositeCommit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_in: ["_C"]}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "_C"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameInEmpty_ReturnsNoCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_in: []}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameNotIn_ExcludesMatchingCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_nin: ["_C", "age"]}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameNotInComposite_ExcludesCompositeCommit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_nin: ["_C"]}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "age"},
						{"fieldName": "name"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameNotInEmpty_ReturnsAllCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_nin: []}}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "age"},
						{"fieldName": "name"},
						{"fieldName": "_C"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameInAndCondition_ReturnsFilteredCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {_and: [{fieldName: {_in: ["age", "name"]}}, {fieldName: {_neq: "age"}}]}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameNotInOrCondition_ReturnsFilteredCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					_commits(filter: {_or: [{fieldName: {_nin: ["_C", "name"]}}, {fieldName: {_eq: "_C"}}]}) {
						fieldName
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "age"},
						{"fieldName": "_C"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
