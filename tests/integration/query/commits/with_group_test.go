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

func TestQueryCommitsWithGroupBy(t *testing.T) {
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
				Request: ` {
						_commits(groupBy: [height]) {
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"height": int64(2),
						},
						{
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
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
				Request: ` {
						_commits(groupBy: [height]) {
							height
							GROUP {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"height": int64(2),
							"GROUP": []map[string]any{
								{

									"cid": testUtils.ValidCID(),
								},
								{
									"cid": testUtils.ValidCID(),
								},
							},
						},
						{
							"height": int64(1),
							"GROUP": []map[string]any{
								{

									"cid": testUtils.ValidCID(),
								},
								{
									"cid": testUtils.ValidCID(),
								},
								{
									"cid": testUtils.ValidCID(),
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

// This is an odd test, but we need to make sure it works
func TestQueryCommitsWithGroupByCidWithChild(t *testing.T) {
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
				Request: ` {
						_commits(groupBy: [cid]) {
							cid
							GROUP {
								height
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{

							"cid": testUtils.ValidCID(),
							"GROUP": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{

							"cid": testUtils.ValidCID(),
							"GROUP": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{

							"cid": testUtils.ValidCID(),
							"GROUP": []map[string]any{
								{
									"height": int64(1),
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

func TestQueryCommitsWithGroupByDocID(t *testing.T) {
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
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc: `{
					"age":	26
				}`,
			},
			&action.Request{
				Request: ` {
						_commits(groupBy: [docID]) {
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"docID": "{{.DocID0_0}}",
						},
						{
							"docID": "{{.DocID0_1}}",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldName(t *testing.T) {
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
				Request: ` {
						_commits(groupBy: [fieldName]) {
							fieldName
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
						},
						{
							"fieldName": "name",
						},
						{
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldNameWithChild(t *testing.T) {
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
				Request: ` {
						_commits(groupBy: [fieldName]) {
							fieldName
							GROUP {
								height
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"GROUP": []map[string]any{
								{
									"height": int64(2),
								},
								{
									"height": int64(1),
								},
							},
						},
						{
							"fieldName": "name",
							"GROUP": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"fieldName": "_C",
							"GROUP": []map[string]any{
								{
									"height": int64(2),
								},
								{
									"height": int64(1),
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
