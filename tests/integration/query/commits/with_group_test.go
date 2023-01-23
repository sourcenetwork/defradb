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
						"cid": "bafybeicvef4ugls2dl7j4hibt2ahxss2i2i4bbgps7tkjiaoybp6q73mca",
					},
					{
						"cid": "bafybeica4js2abwqjjrz7dcialbortbz32uxp7ufxu7yljbwvmhjqqxzny",
					},
				},
			},
			{
				"height": int64(1),
				"_group": []map[string]any{
					{
						"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
					},
					{
						"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
					},
					{
						"cid": "bafybeih75uved5tzh5p7q4kzmkal7ezpjph7fnij6vn5krkgyc3jap2k5m",
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
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
				"_group": []map[string]any{
					{
						"height": int64(1),
					},
				},
			},
			{
				"cid": "bafybeih75uved5tzh5p7q4kzmkal7ezpjph7fnij6vn5krkgyc3jap2k5m",
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
