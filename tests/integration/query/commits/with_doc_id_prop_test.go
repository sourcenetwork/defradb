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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDProperty(t *testing.T) {
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
			&action.Request{
				Request: `query {
						_commits {
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"docID": "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
						},
						{
							"docID": "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
						},
						{
							"docID": "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
