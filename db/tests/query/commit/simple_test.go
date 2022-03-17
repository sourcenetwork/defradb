// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commit

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQueryOneCommit(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "query for a single block by CID",
		Query: `query {
					commit(cid: "bafybeih2egliqqrwwykitohimsvqgtg4dvx5ts5vraadhjdz5ls2cnvpnq") {
						cid
						height
						delta
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"cid":    "bafybeih2egliqqrwwykitohimsvqgtg4dvx5ts5vraadhjdz5ls2cnvpnq",
				"height": int64(1),
				// cbor encoded delta
				"delta": []uint8{0xa2, 0x63, 0x41, 0x67, 0x65, 0x15, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x64, 0x4a, 0x6f, 0x68, 0x6e},
			},
		},
	}

	executeTestCase(t, test)
}
