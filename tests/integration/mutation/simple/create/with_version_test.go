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
	test := testUtils.RequestTestCase{
		Description: "Simple create mutation",
		Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_version {
							cid
						}
					}
				}`,
		Results: []map[string]any{
			{
				"_version": []map[string]any{
					{
						"cid": "bafybeif5xonyzwmg5y5ocebvjkb4vs3i3qmrnuwwtf4yshvabqcqcxnwky",
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}
