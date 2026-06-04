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

package crdt

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldCRDTObjectWithBoolFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
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

func TestCollectionVersionUpdatesAddFieldCRDTObjectWithBoolFieldErrorsMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
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
