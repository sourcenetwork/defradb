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

func TestQueryCommits(t *testing.T) {
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
				Results: []map[string]any{
					{
						"cid": "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
					},
					{
						"cid": "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
					},
					{
						"cid": "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
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
				Results: []map[string]any{
					{
						"cid": "bafyreiazgtllwk7znzuapv3fsukzhpekqqjjvgv4fzypkfp7mljfabie3q",
					},
					{
						"cid": "bafyreicbr2jo7y4d6773q66kxvzq4k3jss2rw5ysr3co2mjdhcdyiz7buq",
					},
					{
						"cid": "bafyreihmvuytwy5ofcm5bqyazxwnquksutxvybznavmw23vddb7nooh6pq",
					},
					{
						"cid": "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
					},
					{
						"cid": "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
					},
					{
						"cid": "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
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
				Results: []map[string]any{
					{
						"cid":             "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
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
				Results: []map[string]any{
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
						"fieldName": nil,
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

func TestQueryCommitsWithFieldIDField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
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
							fieldId
						}
					}
				`,
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

func TestQueryCommitsWithFieldIDFieldWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
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
							fieldId
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
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

func TestQuery_CommitsWithAllFieldsWithUpdate_NoError(t *testing.T) {
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
							collectionID
							delta
							docID
							fieldId
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":          "bafyreigurfgpfvcm4uzqxjf4ur3xegxbebn6yoogjrvyaw6x7d2ji6igim",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreif632ewkphjjwxcthemgbkgtm25faw22mvw7eienu5gnazrao33ba",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
								"name": "_head",
							},
							{
								"cid":  "bafyreigurfgpfvcm4uzqxjf4ur3xegxbebn6yoogjrvyaw6x7d2ji6igim",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafyreichxrfyhajs7rzp3wh5f2zrmt3zkjqan5dmxoy4qno5ozy7omzfpq",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreidksmyoo6txzmcygby6quhdkzymlqoaxpg75ehlxjdneotjzbih6y",
								"name": "name",
							},
							{
								"cid":  "bafyreietqxguz3xlady4gfaqnbeamwsnrwfkufykkpprxej7a77ba7siay",
								"name": "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
