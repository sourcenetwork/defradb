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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
								"cid": "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
							},
							{
								"cid": "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
							},
						},
					},
					{
						"height": int64(1),
						"_group": []map[string]any{
							{
								"cid": "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
							},
							{
								"cid": "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
							},
							{
								"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid": "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithGroupByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by dockey",
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
						commits(groupBy: [dockey]) {
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"dockey": "bae-72f3dc53-1846-55d5-915c-28c4e83cc891",
					},
					{
						"dockey": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
