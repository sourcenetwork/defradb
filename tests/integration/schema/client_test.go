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

	schemaTypes "github.com/sourcenetwork/defradb/request/graphql/schema/types"
)

// TestIntrospectionExplainTypeDefined tests that the introspection query returns a schema that
// defines the ExplainType enum.
func TestIntrospectionExplainTypeDefined(t *testing.T) {
	test := RequestTestCase{
		Schema: []string{},
		IntrospectionRequest: `
			query IntrospectionQuery {
				__schema {
					types {
						kind
						name
						description
					}
				}
			}
		`,
		ContainsData: map[string]any{
			"__schema": map[string]any{
				"types": []any{
					map[string]any{
						"description": schemaTypes.ExplainEnum.Description(),
						"kind":        "ENUM",
						"name":        "ExplainType",
					},
				},
			},
		},
	}

	ExecuteRequestTestCase(t, test)
}
