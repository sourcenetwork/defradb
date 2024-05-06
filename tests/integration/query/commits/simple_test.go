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
						"cid": "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
					},
					{
						"cid": "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
					},
					{
						"cid": "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
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
						"cid": "bafybeihnalsemihsyycy3vaxbhq6iqrixmsk5k3idq52u76h2f5wkvobx4",
					},
					{
						"cid": "bafybeifxk5rhzuemqn2o35hh7346gydqlfmhkdzeguiqo5vczgyz4xz7rm",
					},
					{
						"cid": "bafybeig36zwhejk54nvvey5wsfbl7rzm7xscsyji5uqp6j4hw4zh7dhep4",
					},
					{
						"cid": "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
					},
					{
						"cid": "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
					},
					{
						"cid": "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIdField(t *testing.T) {
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
						"cid":             "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
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
						"cid":          "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
								"name": "_head",
							},
							{
								"cid":  "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
								"name": "age",
							},
							{
								"cid":  "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
