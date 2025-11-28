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

func TestQueryCommitsWithDocIDAndLinkCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							_count(field: links)
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"_count": 0,
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"_count": 0,
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"_count": 2,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
