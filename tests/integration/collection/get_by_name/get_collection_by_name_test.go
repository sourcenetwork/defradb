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

package get_by_name

import (
	"testing"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestGetCollectionByName_Exists_ReturnsCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.GetCollectionByName{
				Name: "Users",
				ExpectedResult: client.CollectionVersion{
					Name:           "Users",
					IsActive:       true,
					IsMaterialized: true,
					Fields: []client.CollectionFieldDescription{
						{
							Name: "_docID",
							Kind: client.FieldKind_DocID,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionByName_DoesNotExist_ReturnsNotFoundError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.GetCollectionByName{
				Name:          "DoesNotExist",
				ExpectedError: client.ErrCollectionNotFound.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionByName_WithNACAndAuthorizedIdentity_ReturnsCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			// The identity must be forwarded by every client for this to succeed.
			&action.GetCollectionByName{
				Identity: testUtils.ClientIdentity(1),
				Name:     "Users",
				ExpectedResult: client.CollectionVersion{
					Name:           "Users",
					IsActive:       true,
					IsMaterialized: true,
					Fields: []client.CollectionFieldDescription{
						{
							Name: "_docID",
							Kind: client.FieldKind_DocID,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionByName_WithNACAndWrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {}
				`,
			},
			&action.GetCollectionByName{
				Identity:      testUtils.ClientIdentity(2),
				Name:          "Users",
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
