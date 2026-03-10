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

package encryption

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestDocEncryptionField_IfFieldDoesNotExistInGQLSchema_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"points"},
				ExpectedError:   "Argument \"encryptFields\" has invalid value [points].",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfAttemptToEncryptBuiltinFieldInGQLSchema_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_docID"},
				ExpectedError:   "Argument \"encryptFields\" has invalid value [_docID].",
			},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_version"},
				ExpectedError:   "Argument \"encryptFields\" has invalid value [_version].",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfFieldDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionSaveMutationType,
			state.CollectionNamedMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"points"},
				ExpectedError:   client.NewErrFieldNotExist("points").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfAttemptToEncryptBuiltinField_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionSaveMutationType,
			state.CollectionNamedMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_docID"},
				ExpectedError:   db.NewErrCanNotEncryptBuiltinField("_docID").Error(),
			},
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_version"},
				ExpectedError:   client.NewErrFieldNotExist("_version").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
