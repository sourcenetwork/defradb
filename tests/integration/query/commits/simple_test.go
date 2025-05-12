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

	"github.com/onsi/gomega"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	headCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Description: "Simple all commits query",
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
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": gomega.And(nameCid, uniqueCid),
						},
						{
							"cid": gomega.And(ageCid, uniqueCid),
						},
						{
							"cid": gomega.And(headCid, uniqueCid),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, multiple docs",
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
						"name":	"Shahzad",
						"age":	28
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreigyslrbac5tgyapdtdo5jrqwon4hewndxjztlbenx632zgmqsma2y",
						},
						{
							"cid": "bafyreiemjaudcun4zej2bampuglyp5ad7d7a3imdizf7ko7stccafyvj44",
						},
						{
							"cid": "bafyreigosblls6ehat2x5tbwbkkqds5yfog6zkwuq7655pljftuy5vj5ke",
						},
						{
							"cid": "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
						},
						{
							"cid": "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIDField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding schemaVersionId",
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
				Request: `query {
						commits {
							cid
							schemaVersionId
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":             "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"schemaVersionId": "bafyreigk2gtae2irmijtkb7z736r3lpssqv7cvmbrp3p6x6ouw7nakc4nm",
						},
						{
							"cid":             "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"schemaVersionId": "bafyreigk2gtae2irmijtkb7z736r3lpssqv7cvmbrp3p6x6ouw7nakc4nm",
						},
						{
							"cid":             "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"schemaVersionId": "bafyreigk2gtae2irmijtkb7z736r3lpssqv7cvmbrp3p6x6ouw7nakc4nm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
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

func TestQueryCommitsWithFieldNameFieldAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "age",
						},
						{
							"fieldName": "age",
						},
						{
							"fieldName": "name",
						},
						{
							"fieldName": "_C",
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

func TestQuery_CommitsWithAllFieldsWithUpdate_NoError(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	ageUpdateCid := testUtils.NewSameValue()
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	updateCompositeCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
						"age":	22
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								name
							}
							signature {
								type
						}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":       gomega.And(ageUpdateCid, uniqueCid),
							"delta":     testUtils.CBORValue(22),
							"docID":     "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							"fieldName": "age",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":  ageCreateCid,
									"name": "_head",
								},
							},
							"signature": nil,
						},
						{
							"cid":       gomega.And(ageCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue(21),
							"docID":     "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
							"signature": nil,
						},
						{
							"cid":       gomega.And(nameCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue("John"),
							"docID":     "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
							"signature": nil,
						},
						{
							"cid":       gomega.And(updateCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							"fieldName": "_C",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":  createCompositeCid,
									"name": "_head",
								},
								{
									"cid":  ageUpdateCid,
									"name": "age",
								},
							},
							"signature": nil,
						},
						{
							"cid":       gomega.And(createCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":  ageCreateCid,
									"name": "age",
								},
								{
									"cid":  nameCreateCid,
									"name": "name",
								},
							},
							"signature": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithAlias_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits with alias query",
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
				Request: `query {
					history: commits {
						cid
					}
				}`,
				Results: map[string]any{
					"history": []map[string]any{
						{
							"cid": "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
						},
						{
							"cid": "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
