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

func TestQueryCommitsWithDepth1(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// "Age" field head
							"cid":    uniqueCid,
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    uniqueCid,
							"height": int64(1),
						},
						{
							"cid":    uniqueCid,
							"height": int64(2),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(depth: 2) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// Composite head
							"cid":    uniqueCid,
							"height": int64(3),
						},
						{
							// Composite head -1
							"cid":    uniqueCid,
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    uniqueCid,
							"height": int64(1),
						},
						{
							// "Age" field head
							"cid":    uniqueCid,
							"height": int64(3),
						},
						{
							// "Age" field head -1
							"cid":    uniqueCid,
							"height": int64(2),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameAndDepth_ReturnsCommitsAtAllHeights(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				Doc: `{"age": 22}`,
			},
			&action.UpdateDoc{
				Doc: `{"age": 23}`,
			},
			&action.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: "age"}}, depth: 2) {
							fieldName
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"height":    int64(3),
						},
						{
							"fieldName": "age",
							"height":    int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
