// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

func TestMutationCreateSimpleReturnVersionCID(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple create mutation",
		Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_version {
							cid
						}
					}
				}`,
		Results: []map[string]interface{}{
			{
				"_version": []map[string]interface{}{
					{
						"cid": "bafybeihsaeu7o2kep75fadotbqurrvqnamkjqr6cnpyvxxb3iolzxvzxve",
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}
