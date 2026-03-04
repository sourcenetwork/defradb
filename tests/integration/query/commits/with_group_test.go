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
			testUtils.UpdateDoc{
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
			testUtils.UpdateDoc{
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
									"cid": "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
								},
								{
									"cid": "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
								},
							},
						},
						{
							"height": int64(1),
							"GROUP": []map[string]any{
								{
									"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
								},
								{
									"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
								},
								{
									"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
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
							"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"GROUP": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"GROUP": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"GROUP": []map[string]any{
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
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
							"docID": "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
						},
						{
							"docID": "bae-2487fd12-227f-582b-a7ed-3dd5d4b61fce",
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
			testUtils.UpdateDoc{
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
			testUtils.UpdateDoc{
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
