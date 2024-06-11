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
			// _docID: "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
			Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-819c9c03-9d49-5fd5-aaee-0dc5a70bbe44"
				}`,
		},
		{
			CollectionID: 2,
			// _docID: "bae-68cb395d-df73-5bcb-b623-615a140dee12"
			Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-9bf0272a-c521-5bef-a7ba-642e8be6e433"
				}`,
		},
	}
}

func create2AuthorContactDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 3,
			// "author_id": "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
			// _docID: "bae-819c9c03-9d49-5fd5-aaee-0dc5a70bbe44"
			Doc: `{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"address_id": "bae-14f20db7-3654-58de-9156-596ef2cfd790"
				}`,
		},
		{
			CollectionID: 3,
			// "author_id": "bae-68cb395d-df73-5bcb-b623-615a140dee12",
			// _docID: "bae-9bf0272a-c521-5bef-a7ba-642e8be6e433"
			Doc: `{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"address_id": "bae-49f715e7-7f01-5509-a213-ed98cb81583f"
				}`,
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
