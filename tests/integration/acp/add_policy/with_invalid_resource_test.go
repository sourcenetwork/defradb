// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_add_policy

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_OneResourceThatIsEmpty_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, one resource that is empty, should return error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

				Policy: `
                    name: a policy
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                `,

				ExpectedError: "resource users: resource missing owner relation: invalid policy",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
