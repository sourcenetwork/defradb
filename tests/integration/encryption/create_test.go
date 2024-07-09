// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestDocEncryptionField_IfFieldDoesNotExistInGQLSchema_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"points"},
				ExpectedError:   "Argument \"encryptFields\" has invalid value [points].",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfAttemptToEncryptBuildInFieldInGQLSchema_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_docID"},
				ExpectedError:   "Argument \"_docID\" has invalid value [points].",
			},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_version"},
				ExpectedError:   "Argument \"_version\" has invalid value [points].",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfFieldDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"points"},
				ExpectedError:   client.NewErrFieldNotExist("points").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_IfAttemptToEncryptBuildInField_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_docID"},
				ExpectedError:   db.NewErrCanNotEncryptBuiltinField("_docID").Error(),
			},
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"_version"},
				ExpectedError:   client.NewErrFieldNotExist("_version").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
