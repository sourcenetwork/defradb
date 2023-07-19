// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test is for documentation reasons only. This is not
// desired behaviour (should return all latest commits).
func TestQueryLatestCommits(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query",
		Request: `query {
					latestCommits {
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
					"name": "John",
					"age": 21
				}`,
			},
		},
		ExpectedError: "Field \"latestCommits\" argument \"dockey\" of type \"ID!\" is required but not provided.",
	}

	executeTestCase(t, test)
}
