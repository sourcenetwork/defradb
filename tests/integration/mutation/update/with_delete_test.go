// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestUpdateSave_DeletedDoc_DoesNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Save existing, deleted document",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// We only wish to test collection.Save in this test.
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
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
