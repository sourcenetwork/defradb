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
				"author_id": testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":      "To my dear readers",
				"pages":     200,
				"author_id": testUtils.NewDocIndex(2, 1),
			},
		},
		{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":      "Twinklestar's Favourite Xmas Cookie",
				"pages":     300,
				"author_id": testUtils.NewDocIndex(2, 1),
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
				"author_id":    testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":         "A Time for Mercy",
				"pages":        333,
				"chapterPages": []int64{0, 22, 101, 321},
				"author_id":    testUtils.NewDocIndex(2, 0),
			},
		},
		{
			CollectionID: 1,
			DocMap: map[string]any{
				"name":      "Theif Lord",
				"pages":     20,
				"author_id": testUtils.NewDocIndex(2, 1),
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
				"contact_id": testUtils.NewDocIndex(3, 0),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Cornelia Funke",
				"age":        62,
				"verified":   false,
				"contact_id": testUtils.NewDocIndex(3, 1),
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
				"address_id": testUtils.NewDocIndex(4, 0),
			},
		},
		{
			CollectionID: 3,
			DocMap: map[string]any{
				"cell":       "5197212302",
				"email":      "cornelia_funke@example.com",
				"address_id": testUtils.NewDocIndex(4, 1),
			},
		},
	}
}

func create2AddressDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 4,
			// _docID: bae-14f20db7-3654-58de-9156-596ef2cfd790
			Doc: `{
					"city": "Waterloo",
					"country": "Canada"
				}`,
		},
		{
			CollectionID: 4,
			// _docID: bae-49f715e7-7f01-5509-a213-ed98cb81583f
			Doc: `{
					"city": "Brampton",
					"country": "Canada"
				}`,
		},
	}
}
