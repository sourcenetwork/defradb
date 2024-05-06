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

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown document ID",
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
						commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, with links",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
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

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results",
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
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
					},
					{
						"cid":    "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results and links",
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
				Request: `query {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"links": []map[string]any{
							{
								"cid":  "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
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
						"cid": "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
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
