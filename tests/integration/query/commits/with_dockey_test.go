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
				"cid": "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
			},
			{
				"cid": "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
			},
			{
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
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
				"cid":   "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"links": []map[string]any{},
			},
			{
				"cid":   "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"links": []map[string]any{},
			},
			{
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"links": []map[string]any{
					{
						"cid":  "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
						"name": "Age",
					},
					{
						"cid":  "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
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
				"cid":    "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
				"height": int64(2),
			},
			{
				"cid":    "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"height": int64(1),
			},
			{
				"cid":    "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"height": int64(1),
			},
			{
				"cid":    "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
				"height": int64(2),
			},
			{
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
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
				"cid": "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
				"links": []map[string]any{
					{
						"cid":  "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
						"name": "_head",
					},
				},
			},
			{
				"cid":   "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"links": []map[string]any{},
			},
			{
				"cid":   "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"links": []map[string]any{},
			},
			{
				"cid": "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
				"links": []map[string]any{
					{
						"cid":  "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
						"name": "Age",
					},
					{
						"cid":  "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
						"name": "_head",
					},
				},
			},
			{
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"links": []map[string]any{
					{
						"cid":  "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
						"name": "Age",
					},
					{
						"cid":  "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
