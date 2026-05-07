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

package remove

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionRemove_CollectionWithConcurrentAddField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Async{
				Child: &action.PatchCollection{
					Patch: `
						[
							{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
						]
					`,
					// If this action completes last, it will transaction error, this is desirable.  Any other
					// error is unwanted.
					//
					// todo - sometimes this is `add operation does not apply: doc is missing path: "/Users/Fields/-": missing value`
					// (is correct if `remove` completely completes before this action starts)
					IgnoreError: "transaction conflict",
				},
			},
			&action.Async{
				Child: &action.PatchCollection{
					Patch: `
						[
							{
								"op": "remove",
								"path": "/Users"
							}
						]
					`,
					// If this action completes last, it will transaction error, this is desirable.  Any other
					// error is unwanted.
					IgnoreError: "transaction conflict",
				},
			},
			&action.Await{},
			// WARNING - There is still a test gap here, we should check that exactly one of the patch actions returned
			// a transaction conflict error.  Longer term, we probably want no errors, but then we need to assert that the
			// possible end states are valid, which they might not be at the moment if not for the transaction conflict errors.
			//
			// https://github.com/sourcenetwork/defradb/issues/4510
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
