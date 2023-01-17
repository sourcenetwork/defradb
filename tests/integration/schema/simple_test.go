// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"
)

func TestSchemaSimpleCreatesSchemaGivenEmptyType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedData: map[string]any{
			"__type": map[string]any{
				"name": "users",
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchema(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedError: "schema type already exists",
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenNewTypes(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
			`
				type books {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "books") {
					name
				}
			}
		`,
		ExpectedData: map[string]any{
			"__type": map[string]any{
				"name": "books",
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaWithDefaultFieldsGivenEmptyType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
					fields {
						name
						type {
						  name
						  kind
						}
					}
				}
			}
		`,
		ExpectedData: map[string]any{
			"__type": map[string]any{
				"name":   "users",
				"fields": defaultFields.tidy(),
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenTypeWithInvalidFieldType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {
					Name: NotAType
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedError: "no type found for given name",
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenTypeWithStringField(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {
					Name: String
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
					fields {
						name
						type {
						  name
						  kind
						}
					}
				}
			}
		`,
		ExpectedData: map[string]any{
			"__type": map[string]any{
				"name": "users",
				"fields": defaultFields.append(
					field{
						"name": "Name",
						"type": map[string]any{
							"kind": "SCALAR",
							"name": "String",
						},
					},
				).tidy(),
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}
