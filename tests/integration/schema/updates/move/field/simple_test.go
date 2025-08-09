// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesMoveFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "move", "from": "/Users/Fields/1", "path": "/Users/Fields/-" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: name, ProposedIndex: 1, ExistingIndex: 2",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesMoveFieldErrorsMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "move", "from": "/Users/Fields/1", "path": "/Users/Fields/-" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: name, ProposedIndex: 1, ExistingIndex: 2\nmoving fields is not currently supported. Name: email, ProposedIndex: 2, ExistingIndex: 1",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
