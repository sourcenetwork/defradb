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
						"cid": "bafybeib5nodgdzwhsrnwe6e4b56riltvtru6ai6ipyogrc6ilhczevjq4e",
					},
					{
						"cid": "bafybeienj3ehxysao3xuhrsamnlgs7b4d7p24fsygg5stw3ckj4tevtr34",
					},
				},
			},
			{
				"height": int64(1),
				"_group": []map[string]any{
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
				"cid": "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
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
