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

func TestCollectionVersionUpdatesAddFieldCRDTInvalidErrors(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ":99} }
					]
				`,
				ExpectedError: "CRDT type not supported. Name: foo, CRDTType: unknown",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldCRDTInvalidErrorsMultiple(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ":99} },
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "bar", "Kind": 2, "Typ":99} }
					]
				`,
				ExpectedError: "CRDT type not supported. Name: foo, CRDTType: unknown\nCRDT type not supported. Name: bar",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
