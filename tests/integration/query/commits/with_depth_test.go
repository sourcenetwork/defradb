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

func TestQueryCommitsWithDepth1(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
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
						commits(depth: 1) {
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

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
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
						commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// "Age" field head
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 2) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// Composite head
						"cid":    "bafybeigzaxekosbmrfrzjhkztodipzmz3voiqnia275347b6vkq5keouf4",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeifwa5vgfvnrdwzqmojsxilwbg2k37axh2fs57zfmddz3l5yivn4la",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
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
			testUtils.Request{
				Request: `query {
						commits(depth: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeicacj5fmr267b6kkmv4ck3g5cm5odca7hu7ajwagfttpspbsu7n5u",
					},
					{
						"cid": "bafybeiexu7xpwhyo2azo2ap2nbny5d4chhr725xrhmxnt5ebabucyjlfqu",
					},
					{
						"cid": "bafybeibbp6jn7y2t6jakbdtvboruieo3iobyuumppbwbw7rwkmz4tdh5yq",
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
