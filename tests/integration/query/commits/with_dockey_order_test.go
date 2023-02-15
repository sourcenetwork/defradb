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
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order height desc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: DESC}) {
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
				"cid":    "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
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
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order height asc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
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
				"cid":    "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"height": int64(1),
			},
			{
				"cid":    "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"height": int64(1),
			},
			{
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"height": int64(1),
			},
			{
				"cid":    "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
				"height": int64(2),
			},
			{
				"cid":    "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order cid desc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: DESC}) {
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
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
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
				"cid":    "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"height": int64(1),
			},
			{
				"cid":    "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
				"height": int64(2),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, order cid asc",
		Request: `query {
					commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: ASC}) {
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
				"cid":    "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
				"height": int64(2),
			},
			{
				"cid":    "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"height": int64(1),
			},
			{
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"height": int64(1),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
		Request: `query {
                     commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
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
					`{
                         "Age": 23
                     }`,
					`{
                         "Age": 24
                     }`,
				},
			},
		},
		Results: []map[string]any{
			{
				"cid":    "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"height": int64(1),
			},
			{
				"cid":    "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"height": int64(1),
			},
			{
				"cid":    "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"height": int64(1),
			},
			{
				"cid":    "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
				"height": int64(2),
			},
			{
				"cid":    "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
				"height": int64(2),
			},
			{
				"cid":    "bafybeibgqpp5sovdh73u7bhi33ib5tlz23evscveh5weflpxwkfn3ayaiq",
				"height": int64(3),
			},
			{
				"cid":    "bafybeifm6qsukak7jnyqtymx2q4viehpxhbbdkwja6eysa2cflfhbeeeaq",
				"height": int64(3),
			},
			{
				"cid":    "bafybeiek27ivyfznotufsb43tshjrinzq7yx2adt2kbwdvj3e7fw6odxeu",
				"height": int64(4),
			},
			{
				"cid":    "bafybeibhiyqcprl4uagkgzxwpyeq4elbf5cpi2xqkvkz5if66esb2jrrcy",
				"height": int64(4),
			},
		},
	}

	executeTestCase(t, test)
}
