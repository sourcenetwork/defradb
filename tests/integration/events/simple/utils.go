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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	testUtilsEvt "github.com/sourcenetwork/defradb/tests/integration/events"
)

var schema = `
	type Users {
		name: String
	}
`

var colDefMap = make(map[string]client.CollectionDefinition)

func init() {
	c, err := testUtils.ParseSDL(schema)
	if err != nil {
		panic(err)
	}
	colDefMap = c
}

func executeTestCase(t *testing.T, test testUtilsEvt.TestCase) {
	testUtilsEvt.ExecuteRequestTestCase(t, schema, test)
}
