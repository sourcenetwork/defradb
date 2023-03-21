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

const clientIntrospectionRequest string = `
query IntrospectionQuery {
	__schema {
	queryType { name }
	mutationType { name }
	subscriptionType { name }
	types {
		...FullType
	}
	directives {
		name
		description
		locations
		args {
		...InputValue
		}
	}
	}
}

fragment FullType on __Type {
	kind
	name
	description
	fields(includeDeprecated: true) {
	name
	description
	args {
		...InputValue
	}
	type {
		...TypeRef
	}
	isDeprecated
	deprecationReason
	}
	inputFields {
	...InputValue
	}
	interfaces {
	...TypeRef
	}
	enumValues(includeDeprecated: true) {
	name
	description
	isDeprecated
	deprecationReason
	}
	possibleTypes {
	...TypeRef
	}
}

fragment InputValue on __InputValue {
	name
	description
	type { ...TypeRef }
	defaultValue
}

fragment TypeRef on __Type {
	kind
	name
	ofType {
	kind
	name
	ofType {
		kind
		name
		ofType {
		kind
		name
		ofType {
			kind
			name
			ofType {
			kind
			name
			ofType {
				kind
				name
				ofType {
				kind
				name
				}
			}
			}
		}
		}
	}
	}
}
  `

// TestClientIntrospectionExplainTypeDefined tests that the introspection query returns a schema that
// defines the ExplainType enum.
func TestClientIntrospectionExplainTypeDefined(t *testing.T) {
	test := RequestTestCase{
		Schema:               []string{},
		IntrospectionRequest: clientIntrospectionRequest,
		ContainsData: map[string]any{
			"__schema": map[string]any{
				"types": []any{
					map[string]any{
						"description": "",
						"enumValues": []any{
							map[string]any{
								"deprecationReason": nil,
								"description":       "Simple explaination - dump of the plan graph.",
								"isDeprecated":      false,
								"name":              "simple",
							},
						},
						"fields":        nil,
						"inputFields":   nil,
						"interfaces":    nil,
						"kind":          "ENUM",
						"name":          "ExplainType",
						"possibleTypes": nil,
					},
				},
			},
		},
	}

	ExecuteRequestTestCase(t, test)
}
