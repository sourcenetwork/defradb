// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package complex

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/testing/db"
)

func TestQueryComplex(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "multinode: One-to-one relation query with no filter",
		Query: `query {
			book {
				name
				author {
					name
				}
				publisher {
					name
				}
			}
		}`,
		Docs: map[int][]string{
			//books
			0: {
				// bae-7e5ae688-3a77-5b4f-a74c-59301bd1eb25
				(`{
					"name": "The Coffee Table Book",
					"rating": 4.9,
					"publisher_id": "bae-81804a20-4d08-509e-a3e8-fd770622a356"
				}`)},
			//authors
			1: {
				// bae-5eae6a8a-0c52-535c-9c20-df42b7044e20
				(`{
					"name": "Cosmo Kramer",
					"age": 44,
					"verified": true,
					"wrote_id": "bae-7e5ae688-3a77-5b4f-a74c-59301bd1eb25"
				}`)},
			// publishers
			2: {
				// bae-81804a20-4d08-509e-a3e8-fd770622a356
				(`{
					"name": "Pendant Publishing",
					"address": "600 Madison Ave., New York, New York"
				}`)},
		},
		Results: []map[string]interface{}{
			{
				"name": "The Coffee Table Book",
				"author": map[string]interface{}{
					"name": "Cosmo Kramer",
				},
				"publisher": map[string]interface{}{
					"name": "Pendant Publishing",
				},
			},
		},
	}

	executeTestCase(t, test)
}
