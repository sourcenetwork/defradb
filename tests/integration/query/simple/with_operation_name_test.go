package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimpleMultipleOperationsWithOperationName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query multiple operations with operation name",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
							"Name": "Bob",
						},
						{
							"Name": "Alice",
						},
					},
				},
			},
			testUtils.Request{
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
							"Age": int64(21),
						},
						{
							"Age": int64(40),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleMultipleOperationsWithNoOperationName_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query multiple operations with no operation name",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
