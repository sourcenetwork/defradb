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

func TestQueryCommitsWithUnknownDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown dockey",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.Request{
				Request: `query {
						commits(dockey: "unknown dockey") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
						}
					}`,
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, with links",
		Actions: []any{
			updateUserCollectionSchema(),
			createJohnDoc(),
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
								"name": "Age",
							},
							{
								"cid":  "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
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
						"cid":    "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"height": int64(2),
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results and links",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeidxeexqpsbf2qqrrkrysdztf2q5mqfwabwrcxdkjuolf6fsyzzyh4",
						"links": []map[string]any{
							{
								"cid":  "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
								"name": "Age",
							},
							{
								"cid":  "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
								"name": "_head",
							},
						},
					},
					{
						"cid": "bafybeidcatznm2mlsymcytrh5fkpdrazensg5fsvn2uavcgiq2bf26lzey",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
								"name": "Age",
							},
							{
								"cid":  "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
