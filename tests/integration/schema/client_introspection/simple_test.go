// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
	testUtils.ExecuteTEMP(t, test)
}
