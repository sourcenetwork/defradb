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

func TestQueryCommitsWithDockeyWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey and typename",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							__typename
						}
					}`,
				Results: []map[string]any{
					{
						"cid":        "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"__typename": "Commit",
					},
					{
						"cid":        "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"__typename": "Commit",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
