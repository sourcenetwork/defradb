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

func TestQueryCommits(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query",
		Request: `query {
					commits {
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
				"cid": "bafybeid2b6a5vbqzxyxrzvwvkakqlzgcdpcdpkpmufthy4hnasu4zcyzua",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, multiple docs",
		Request: `query {
					commits {
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
				"cid": "bafybeibnwzoekenlil5sdltgmkmfngoifd5uwcpgzzry7rokrd5qurgdve",
			},
			{
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
			{
				"cid": "bafybeihhypcsqt7blkrqtcmpl43eo3yunrog5pchox5naji6hisdme4swm",
			},
			{
				"cid": "bafybeid2b6a5vbqzxyxrzvwvkakqlzgcdpcdpkpmufthy4hnasu4zcyzua",
			},
		},
	}

	executeTestCase(t, test)
}
