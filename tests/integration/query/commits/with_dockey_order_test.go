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

func TestQueryCommitsWithDockeyAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height desc",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 23
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
						"Age": 24
					}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeifaxl4u5wmokgr4jviru6dz7teg7f2fomusxrvh7o5nh2a32jk3va",
						"height": int64(3),
					},
					{
						"cid":    "bafybeieh6icybinqmuz7or7nqjt445zu2qbnxccuu64i5jhgonhlj2k5sy",
						"height": int64(3),
					},
					{
						"cid":    "bafybeic4slf53yiert4jrdeyvqij3rnat2iikgbn7fdn6n6zowu66dylbi",
						"height": int64(4),
					},
					{
						"cid":    "bafybeic4rhgl4pvsh3z5rtr5tzm3zzakyivju6mqemq5v4xqxh7otgrjri",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
