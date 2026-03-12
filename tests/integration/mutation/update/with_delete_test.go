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

package update

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestUpdateSave_DeletedDoc_DoesNothing(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// We only wish to test collection.Save in this test.
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"name": "Fred"
				}`,
				ExpectedError: "a document with the given ID has been deleted",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
