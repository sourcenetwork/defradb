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

package client_introspection

import (
	_ "embed"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

//go:embed altair_graphiql_postman_2023.gql
var clientIntrospectionQuery string

func TestClientIntrospectionBasic(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ClientIntrospectionRequest{
				Request: clientIntrospectionQuery,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
