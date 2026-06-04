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

package collection_version

import (
	"testing"

	schemaTypes "github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestIntrospectionExplainTypeDefined tests that the introspection query returns a GQL schema that
// defines the ExplainType enum.
func TestIntrospectionExplainTypeDefined(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.IntrospectionRequest{
				Request: `
					query {
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
								"description": schemaTypes.ExplainEnum().Description(),
								"kind":        "ENUM",
								"name":        "ExplainType",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
