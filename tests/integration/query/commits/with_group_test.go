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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by height",
		Request: `query {
					commits(groupBy: [height]) {
						height
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 22
					}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"height": int64(2),
			},
			{
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by height",
		Request: `query {
					commits(groupBy: [height]) {
						height
						_group {
							cid
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 22
					}`,
				},
			},
		},
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
	}

	executeTestCase(t, test)
}

// This is an odd test, but we need to make sure it works
func TestQueryCommitsWithGroupByCidWithChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by cid",
		Request: `query {
					commits(groupBy: [cid]) {
						cid
						_group {
							height
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
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
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithGroupByDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, group by dockey",
		Request: `query {
					commits(groupBy: [dockey]) {
						dockey
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Fred",
					"Age": 25
				}`,
			},
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 22
					}`,
				},
				1: {
					`{
						"Age": 26
					}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
			},
			{
				"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithOrderedByDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, grouped and ordered by height",
		Request: `query {
					commits(groupBy: [dockey], order: {dockey: DESC}) {
						dockey
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Fred",
					"Age": 25
				}`,
			},
		},
		Results: []map[string]any{
			{
				"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
			},
			{
				"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
			},
		},
	}

	executeTestCase(t, test)
}
