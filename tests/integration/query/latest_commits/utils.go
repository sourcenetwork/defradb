// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCollectionGQLSchema = (`
	type Users {
		name: String
		age: Int
		verified: Boolean
	}
`)

const companiesCollectionGQLSchema = (`
	type Companies {
		name: String
	}
`)

func updateUserCollectionSchema() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
		Schema: userCollectionGQLSchema,
	}
}

func updateCompaniesCollectionSchema() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
		Schema: companiesCollectionGQLSchema,
	}
}

func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteRequestTestCase(t, userCollectionGQLSchema, []string{"Users"}, test)
}
