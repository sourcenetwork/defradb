// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

// This test runs ListLenses inside of a transaction with AddLens, and illustrates that
// the lens is seen by the action.
func TestTxn_ListLenses_InsideTxnWithAddLens_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				TransactionID: immutable.Some(1),
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.ListLenses{
				TransactionID: immutable.Some(1),
				ExpectedLenses: map[string]model.Lens{
					"{{.LensID0}}": {
						Lenses: []model.LensModule{
							{
								Path: lenses.CopyModulePath,
								Arguments: map[string]any{
									"src": "name",
									"dst": "fullName",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs ListLenses inside of a separate transaction from AddLens, and illustrates that
// the lens is not seen by the action.
func TestTxn_ListLenses_InsideTxnWithoutAddLens_NoLenses(t *testing.T) {
	// LevelDB does not support concurrent transactions
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				TransactionID: immutable.Some(1),
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.ListLenses{
				TransactionID:  immutable.Some(2),
				ExpectedLenses: map[string]model.Lens{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
