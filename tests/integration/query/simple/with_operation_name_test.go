// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
