// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestMutationAddWithNonNullVariable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"user": map[string]any{
						"name": "Bob",
					},
				}),
				Request: `mutation($user: [UsersMutationInputArg!]!) {
					add_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
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

func TestMutationAddWithDefaultVariable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation($user: [UsersMutationInputArg!] = {name: "Bob"}) {
					add_Users(input: $user) {
						name
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
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

func TestMutationAdd_WithVariableInJSONObject_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embed: JSON
					}
				`,
			},
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"message": "hello",
				}),
				Request: `mutation($message: String) {
					add_Users(input: {embed: {message: $message}}) {
						embed
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"embed": map[string]any{
								"message": "hello",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithJSONVariable_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						embed: JSON
					}
				`,
			},
			&action.Request{
				Variables: immutable.Some(map[string]any{
					"embed": map[string]any{
						"bar": 1,
					},
				}),
				Request: `mutation($embed: JSON) {
					add_Users(input: {embed: $embed}) {
						embed
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"embed": map[string]any{
								"bar": 1,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
