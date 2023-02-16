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

// This test is for documentation reasons only. This is not
// desired behaviour (should return all commits for dockey-field).
func TestQueryCommitsWithField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with field",
		Request: `query {
					commits (field: "Age") {
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
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryCommitsWithFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with field id",
		Request: `query {
					commits (field: "1") {
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
				"cid": "bafybeiaeic6vhiiw5zu6ju7e47cclvctn6t5pb36fj3mczchyhmctbrr6m",
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryCommitsWithCompositeFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey and field id",
		Request: `query {
					commits(field: "C") {
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
				"cid": "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryCommitsWithCompositeFieldIdWithReturnedSchemaVersionId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query with dockey and field id",
		Request: `query {
					commits(field: "C") {
						cid
						schemaVersionId
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
				"cid":             "bafybeidr2z5ahvvss5j664gxyna5wjil5ndfjbmllnsewkjf6cnsvsmmqu",
				"schemaVersionId": "bafkreibwyhaiseplil6tayn7spazp3qmc7nkoxdjb7uoe5zvcac4pgbwhy",
			},
		},
	}

	executeTestCase(t, test)
}
