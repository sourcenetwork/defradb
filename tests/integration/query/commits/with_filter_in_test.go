// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits_WithFilterFieldNameIn_ReturnsMatchingCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					_commits(filter: {_and: [{fieldName: {_in: ["age", "name"]}}, {fieldName: {_ne: "age"}}]}) {
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
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
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
