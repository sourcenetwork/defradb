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
	test := RequestTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionRequest: `
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

	ExecuteRequestTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchema(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{
			`
				type users {}
			`,
			`
				type users {}
			`,
		},
		IntrospectionRequest: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedError: "schema type already exists",
	}

	ExecuteRequestTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenNewTypes(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{
			`
				type users {}
			`,
			`
				type books {}
			`,
		},
		IntrospectionRequest: `
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

	ExecuteRequestTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaWithDefaultFieldsGivenEmptyType(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{
			`
				type users {}
			`,
		},
		IntrospectionRequest: `
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

	ExecuteRequestTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenTypeWithInvalidFieldType(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{
			`
				type users {
					Name: NotAType
				}
			`,
		},
		IntrospectionRequest: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedError: "no type found for given name",
	}

	ExecuteRequestTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenTypeWithStringField(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{
			`
				type users {
					Name: String
				}
			`,
		},
		IntrospectionRequest: `
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

	ExecuteRequestTestCase(t, test)
}
