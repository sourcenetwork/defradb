// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithMultipleOperationsAndOperationName_ShouldSucceed(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with multiple operations",
		Request: `query usersWithAge {
					Users {
						_docID
						Age
					}
				}
				query usersWithName {
					Users {
						_docID
						Name
					}
				}`,
		OperationName: "usersWithAge",
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"_docID": "bae-d4303725-7db9-53d2-b324-f3ee44020e52",
					"Age":    int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithMultipleOperationsAndNoOperationName_ShouldReturnError(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with multiple operations",
		Request: `query usersWithAge {
					Users {
						_docID
						Age
					}
				}
				query usersWithName {
					Users {
						_docID
						Name
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
		ExpectedError: "request with multiple operations must have an operationName",
	}

	executeTestCase(t, test)
}
