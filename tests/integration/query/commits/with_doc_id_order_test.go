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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
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
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
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
						"cid":    "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
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

func TestQueryCommitsWithDocIDAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid desc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
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

func TestQueryCommitsWithDocIDAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid asc",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
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
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeietdefm4jtgcbpof5elqdptwnuevvzujb3j22n6dytx67vpmb3xn4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihc26puzzgnctvgvgowihytif52tmdj3y2ksx6tvge2wtjkjvtszi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiafufeqwjo5eeeaobwiu6ibf73dnvlefr47mrx45z42337mfnezkq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaho26jaxdjfuvyxozws6ushksjwidllvgai6kgxmqxhzylwzkvte",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiep7c6ouykgidnwzjeasyim3ost5qjkro4qvs62t4u4u7rolbmugm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigzaxekosbmrfrzjhkztodipzmz3voiqnia275347b6vkq5keouf4",
						"height": int64(3),
					},
					{
						"cid":    "bafybeifwa5vgfvnrdwzqmojsxilwbg2k37axh2fs57zfmddz3l5yivn4la",
						"height": int64(3),
					},
					{
						"cid":    "bafybeifn2f5lgzall3dzva47khbtib77lt7ve5qyclou3ihi2hy2uqj4nm",
						"height": int64(4),
					},
					{
						"cid":    "bafybeieijpm36ntafrncl4kgx6dkxgpbftcl4f7obbbmagurcgdoj6sl5y",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
