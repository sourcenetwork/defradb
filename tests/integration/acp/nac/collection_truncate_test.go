// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesCollectionTruncate_AuthorizedIdentity_AllowAccess(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {}
				`,
			},
			// This should work as the identity is authorized.
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(1),
				CollectionIndex: 0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesCollectionTruncate_NoIdentity_NotAuthorizedError(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {}
				`,
			},
			// We haven't authorized non-identities. So, this should error.
			&action.Truncate{
				Identity:        testUtils.NoIdentity(),
				CollectionIndex: 0,
				ExpectedError:   "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesCollectionTruncate_WrongIdentity_NotAuthorizedError(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {}
				`,
			},
			// Wrong user/identity will also not be authorized.
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(2),
				CollectionIndex: 0,
				ExpectedError:   "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
