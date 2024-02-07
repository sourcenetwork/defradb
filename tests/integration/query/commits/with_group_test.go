// Copyright 2022 Democratized Data Foundation
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

func TestQueryCommitsWithGroupBy(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
					},
					{
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
							_group {
								cid
							}
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
						"_group": []map[string]any{
							{
								"cid": "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
							},
							{
								"cid": "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
							},
						},
					},
					{
						"height": int64(1),
						"_group": []map[string]any{
							{
								"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
							},
							{
								"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
							},
							{
								"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
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
		Description: "Simple all commits query, group by cid",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [cid]) {
							cid
							_group {
								height
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"_group": []map[string]any{
							{
								"height": int64(1),
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
		Description: "Simple all commits query, group by document ID",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [docID]) {
							docID
						}
					}`,
				Results: []map[string]any{
					{
						"docID": "bae-72f3dc53-1846-55d5-915c-28c4e83cc891",
					},
					{
						"docID": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [fieldName]) {
							fieldName
						}
					}`,
				Results: []map[string]any{
					{
						"fieldName": "age",
					},
					{
						"fieldName": "name",
					},
					{
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldNameWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [fieldName]) {
							fieldName
							_group {
								height
							}
						}
					}`,
				Results: []map[string]any{
					{
						"fieldName": "age",
						"_group": []map[string]any{
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
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"fieldName": nil,
						"_group": []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [fieldId]) {
							fieldId
						}
					}`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldIDWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
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
			testUtils.Request{
				Request: ` {
						commits(groupBy: [fieldId]) {
							fieldId
							_group {
								height
							}
						}
					}`,
				Results: []map[string]any{
					{
						"fieldId": "1",
						"_group": []map[string]any{
							{
								"height": int64(2),
							},
							{
								"height": int64(1),
							},
						},
					},
					{
						"fieldId": "2",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"fieldId": "C",
						"_group": []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}
