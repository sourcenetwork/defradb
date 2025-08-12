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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldCRDTObjectWithBoolFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ":2} }
					]
				`,
				ExpectedError: "CRDT type not supported. Name: foo, CRDTType: object",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldCRDTObjectWithBoolFieldErrorsMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ":2} },
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "bar", "Kind": 2, "Typ":2} }
					]
				`,
				ExpectedError: "CRDT type not supported. Name: foo, CRDTType: object\nCRDT type not supported. Name: bar",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
