// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesGetNodeIdentity_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates get node identity operation, allow if authorized, otherwise error",
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.GetNodeIdentity{
				Identity:      testUtils.NoIdentity(),
				NodeID:        0,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.GetNodeIdentity{
				Identity:      testUtils.NoIdentity(),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.GetNodeIdentity{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        0,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.GetNodeIdentity{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.GetNodeIdentity{
				Identity:         testUtils.ClientIdentity(1),
				NodeID:           0,
				ExpectedIdentity: testUtils.NodeIdentity(0),
			},
			testUtils.GetNodeIdentity{
				Identity:         testUtils.ClientIdentity(1),
				NodeID:           1,
				ExpectedIdentity: testUtils.NodeIdentity(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
