// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCollectionGQLSchema = (`
	type Users {
		Name: String
		Email: String
		Age: Int
		HeightM: Float
		Verified: Boolean
		CreatedAt: DateTime
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Description:            test.Description,
			SupportedMutationTypes: test.SupportedMutationTypes,
			SupportedClientTypes:   test.SupportedClientTypes,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: userCollectionGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}
