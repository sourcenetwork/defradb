// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package all_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryAllCommits(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query",
		Query: `query {
					allCommits {
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
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
			},
			{
				"cid": "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryAllCommitsMultipleDocs(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query, multiple docs",
		Query: `query {
					allCommits {
						cid
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Shahzad",
					"Age": 28
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeidsa74vl7xvw6tzgt5gmux5ts7lxldxculzgpxl5xura45ckf7e5i",
			},
			{
				"cid": "bafybeibek4lmrb5gtmahgsv33njmk3efty53n7z2rac7fuup7mwpho5zqa",
			},
			{
				"cid": "bafybeidj2zxpwsfbqtny2svdpyan7vu4njfhydzrgpvupv3o673a3dpfku",
			},
			{
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
			},
			{
				"cid": "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
			},
		},
	}

	executeTestCase(t, test)
}
