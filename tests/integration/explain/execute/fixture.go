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
			Doc: `{

					"name": "After Guantánamo, Another Injustice",
					"pages": 100,
					"author_id": "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "To my dear readers",
					"pages": 200,
					"author_id": "bae-68cb395d-df73-5bcb-b623-615a140dee12"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"pages": 300,
					"author_id": "bae-68cb395d-df73-5bcb-b623-615a140dee12"
				}`,
		},
	}
}

func create3BookDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 1,
			Doc: `{
					"name": "Painted House",
					"pages": 78,
					"chapterPages": [1, 22, 33, 44, 55, 66],
					"author_id": "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
				}`,
		},
		{
			CollectionID: 1,
			Doc: `{
					"name": "A Time for Mercy",
					"pages": 333,
					"chapterPages": [0, 22, 101, 321],
					"author_id": "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
				}`,
		},
		{
			CollectionID: 1,
			Doc: `{
					"name": "Theif Lord",
					"pages": 20,
					"author_id": "bae-68cb395d-df73-5bcb-b623-615a140dee12"
				}`,
		},
	}
}

func create2AuthorDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 2,
			// _key: "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
			Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-4db5359b-7dbe-5778-b96f-d71d1e6d0871"
				}`,
		},
		{
			CollectionID: 2,
			// _key: "bae-68cb395d-df73-5bcb-b623-615a140dee12"
			Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-1f19fc5d-de4d-59a5-bbde-492be1757d65"
				}`,
		},
	}
}

func create2AuthorContactDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 3,
			// "author_id": "bae-7f54d9e0-cbde-5320-aa6c-5c8895a89138"
			// _key: "bae-4db5359b-7dbe-5778-b96f-d71d1e6d0871"
			Doc: `{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"address_id": "bae-c8448e47-6cd1-571f-90bd-364acb80da7b"
				}`,
		},
		{
			CollectionID: 3,
			// "author_id": "bae-68cb395d-df73-5bcb-b623-615a140dee12",
			// _key: "bae-1f19fc5d-de4d-59a5-bbde-492be1757d65"
			Doc: `{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"address_id": "bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692"
				}`,
		},
	}
}

func create2AddressDocuments() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 4,
			// "contact_id": "bae-4db5359b-7dbe-5778-b96f-d71d1e6d0871"
			// _key: bae-c8448e47-6cd1-571f-90bd-364acb80da7b
			Doc: `{
					"city": "Waterloo",
					"country": "Canada"
				}`,
		},
		{
			CollectionID: 4,
			// "contact_id": ""bae-1f19fc5d-de4d-59a5-bbde-492be1757d65""
			// _key: bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692
			Doc: `{
					"city": "Brampton",
					"country": "Canada"
				}`,
		},
	}
}
