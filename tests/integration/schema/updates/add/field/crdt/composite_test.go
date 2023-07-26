// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldCRDTCompositeErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with crdt composite (3)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ":3} }
					]
				`,
				ExpectedError: "only default or LWW (last writer wins) CRDT types are supported. Name: foo, CRDTType: 3",
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}
