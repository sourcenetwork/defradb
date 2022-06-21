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

func TestQueryAllCommitsWithDockeyAndUnknownField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with dockey and unknown field",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "not a field") {
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
		Results: []map[string]interface{}{},
	}

	executeTestCase(t, test)
}

func TestQueryAllCommitsWithDockeyAndUnknownFieldId(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with dockey and unknown field id",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "999999") {
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
		Results: []map[string]interface{}{},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (should return all commits for dockey-field).
func TestQueryAllCommitsWithDockeyAndField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with dockey and field",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "Age") {
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
		Results: []map[string]interface{}{},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryAllCommitsWithDockeyAndFieldId(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple all commits query with dockey and field id",
		Query: `query {
					allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "1") {
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
		Results: []map[string]interface{}{
			{
				"cid": "bafybeidst2mzxhdoh4ayjdjoh4vibo7vwnuoxk3xgyk5mzmep55jklni2a",
			},
		},
	}

	executeTestCase(t, test)
}
