// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	testUtilsCol "github.com/sourcenetwork/defradb/tests/integration/collection"
)

var userCollectionGQLSchema = `
	type Users {
		name: String
		age: Int
		heightM: Float
		verified: Boolean
	}
`

var colDefMap = make(map[string]client.CollectionDefinition)

func init() {
	c, err := testUtils.ParseSDL(userCollectionGQLSchema)
	if err != nil {
		panic(err)
	}
	u := c["Users"]
	u.Schema.Root = "bafkreiclkqkxhq3xu3sz5fqcixykk2qfpva5asj3elcaqyxscax66ok4za"
	c["Users"] = u
	colDefMap = c
}

func executeTestCase(t *testing.T, test testUtilsCol.TestCase) {
	testUtilsCol.ExecuteRequestTestCase(t, userCollectionGQLSchema, test)
}
