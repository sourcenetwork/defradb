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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with unknown dockey",
		Request: `query {
					commits(dockey: "unknown dockey") {
						cid
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
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
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
			},
			{
				"cid": "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
			},
			{
				"cid": "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, with links",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						links {
							cid
							name
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
				"cid":   "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
				"links": []map[string]any{},
			},
			{
				"cid":   "bafybeibsaubd2ptp6qqsszv24p73j474amc4pll4oyssnpilofrl575hmy",
				"links": []map[string]any{},
			},
			{
				"cid": "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
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
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, multiple results",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
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
				"cid":    "bafybeigeigzhjtf27o3wkdyq3exmnqhr3npt5psdq3pywpwxxdepiebpdi",
				"height": int64(2),
			},
			{
				"cid":    "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, multiple results and links",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						links {
							cid
							name
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
				"cid": "bafybeigeigzhjtf27o3wkdyq3exmnqhr3npt5psdq3pywpwxxdepiebpdi",
				"links": []map[string]any{
					{
						"cid":  "bafybeihxc6ittcok3rnetguamxfzd3wa534z7zwqsaoppvawu7jx4rdy5u",
						"name": "Age",
					},
					{
						"cid":  "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
						"name": "_head",
					},
				},
			},
			{
				"cid": "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
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
	}

	executeTestCase(t, test)
}
