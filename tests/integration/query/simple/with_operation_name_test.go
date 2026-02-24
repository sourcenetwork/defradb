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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimpleMultipleOperationsWithOperationName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				OperationName: immutable.Some("UsersByName"),
				Request: `query UsersByName {
					Users {
						Name
					}
				}
				query UsersByAge {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
						},
						{
							"Name": "Bob",
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				OperationName: immutable.Some("UsersByAge"),
				Request: `query UsersByName {
					Users {
						Name
					}
				}
				query UsersByAge {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(40),
						},
						{
							"Age": int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleMultipleOperationsWithNoOperationName_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query UsersByName {
					Users {
						Name
					}
				}
				query UsersByAge {
					Users {
						Age
					}
				}`,
				ExpectedError: "Must provide operation name if query contains multiple operations.",
			},
		},
	}

	executeTestCase(t, test)
}
