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

func TestQueryCommitsWithGroupBy(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 22
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
					},
					{
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 22
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
							_group {
								cid
							}
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
						"_group": []map[string]any{
							{
								"cid": "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
							},
							{
								"cid": "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
							},
						},
					},
					{
						"height": int64(1),
						"_group": []map[string]any{
							{
								"cid": "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
							},
							{
								"cid": "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
							},
							{
								"cid": "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

// This is an odd test, but we need to make sure it works
func TestQueryCommitsWithGroupByCidWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by cid",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.Request{
				Request: ` {
						commits(groupBy: [cid]) {
							cid
							_group {
								height
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithGroupByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by dockey",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name": "Fred",
						"Age": 25
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 22
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc: `{
						"Age": 26
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [dockey]) {
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
					{
						"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithOrderedByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, grouped and ordered by height",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "Fred",
					"Age": 25
				}`,
			},
			testUtils.Request{
				Request: ` {
					commits(groupBy: [dockey], order: {dockey: DESC}) {
						dockey
					}
				}`,
				Results: []map[string]any{
					{
						"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
					},
					{
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
