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

package searchable_encryption

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexNew_IfP2PIsDisabled_CanNotDoSEQuery(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			&action.Request{
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				ExpectedError: "Cannot query field \"encrypted_User\" on type \"Query\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
