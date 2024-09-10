package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestMutationCreateWithNonNullVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with non null variable input.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Variables: immutable.Some(map[string]any{
					"user": map[string]any{
						"name": "Bob",
					},
				}),
				Request: `mutation($user: UsersMutationInputArg!) {
					create_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreateWithDefaultVariable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with default variable input.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation($user: UsersMutationInputArg = {name: "Bob"}) {
					create_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
