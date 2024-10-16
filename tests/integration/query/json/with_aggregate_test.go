package json

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithAggregateFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple JSON, aggregate with filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"tree": "maple",
						"age": 250
					}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"tree": "oak",
						"age": 450
					}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					_count(Users: {filter: {custom: {tree: {_eq: "oak"}}}})
				}`,
				Results: map[string]any{
					"_count": 1,
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
