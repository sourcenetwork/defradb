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
	"github.com/sourcenetwork/immutable"
)

func TestColVersionUpdateRemoveView_DoesNotDeadlockIfDeletingVersionWithNoNewFieldsWhilstOtherTxnReading(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: true) {
						name: String
						fullName: String
					}
				`,
			},
			&action.PatchCollection{
				// Patch the view query definition so that it now aliases `name` to `fullName`,
				// this creates a new collection version, without creating any new fields.
				Patch: `
					[
						{
							"op": "replace",
							"path": "/UserView/Query/Query",
							"value": {"Name": "Users", "Fields":[{"Name":"name","Alias":"fullName"}]}
						}
					]
				`,
			},
			&action.Request{
				// Query the UserView using a new transaction in order to acquire a read lock
				TransactionID: immutable.Some(0),
				Request: `query {
					UserView {
						name
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{},
				},
			},
			&action.PatchCollection{
				// Remove the newer version of the view.
				// Because this version did not add any new fields, it is not destroying any local information
				// that cannot be reproduced later (such as field IDs), as such, it does not acquire a write
				// lock.
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreihekbkpnimeo5g5tkv5a3urtcalx2qd447tyhhptjunb7vvpdyvue"
						}
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
