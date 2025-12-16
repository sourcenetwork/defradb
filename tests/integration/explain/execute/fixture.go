// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_execute

import (
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type dataMap = map[string]any

func create3ArticleDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":      "After Guantánamo, Another Injustice",
				"pages":     100,
				"_authorID": testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":      "To my dear readers",
				"pages":     200,
				"_authorID": testUtils.NewDocIndex(2, 1),
			},
		},
		{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":      "Twinklestar's Favourite Xmas Cookie",
				"pages":     300,
				"_authorID": testUtils.NewDocIndex(2, 1),
			},
		},
	}
}

func create3BookDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":         "Painted House",
				"pages":        78,
				"chapterPages": []int64{1, 22, 33, 44, 55, 66},
				"_authorID":    testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":         "A Time for Mercy",
				"pages":        333,
				"chapterPages": []int64{0, 22, 101, 321},
				"_authorID":    testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":      "Theif Lord",
				"pages":     20,
				"_authorID": testUtils.NewDocIndex(2, 1),
			},
		},
	}
}

func create2AuthorDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "John Grisham",
				"age":        65,
				"verified":   true,
				"_contactID": testUtils.NewDocIndex(3, 0),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Cornelia Funke",
				"age":        62,
				"verified":   false,
				"_contactID": testUtils.NewDocIndex(3, 1),
			},
		},
	}
}

func create2AuthorContactDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 3,
			DocMap: map[string]any{
				"cell":       "5197212301",
				"email":      "john_grisham@example.com",
				"_addressID": testUtils.NewDocIndex(4, 0),
			},
		},
		{
			CollectionID: 3,
			DocMap: map[string]any{
				"cell":       "5197212302",
				"email":      "cornelia_funke@example.com",
				"_addressID": testUtils.NewDocIndex(4, 1),
			},
		},
	}
}

func create2AddressDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 4,
			// _docID: bae-186c2484-c3ea-5993-95d6-cb886e1b13a1
			Doc: `{
					"city": "Waterloo",
					"country": "Canada"
				}`,
		},
		{
			CollectionID: 4,
			// _docID: bae-78bc4454-19a6-58ed-9e18-f0ca175dd12c
			Doc: `{
					"city": "Brampton",
					"country": "Canada"
				}`,
		},
	}
}
