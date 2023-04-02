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

func TestQueryLatestCommitsWithDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey",
		Request: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						links {
							cid
							name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
				"links": []map[string]any{
					{
						"cid":  "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"name": "Age",
					},
					{
						"cid":  "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommitsWithDocKeyWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and schema versiion id field",
		Request: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
						cid
						schemaVersionId
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid":             "bafybeiapquwo7dfow7b7ovwrn3nl4e2cv2g5eoufuzylq54b4o6tatfrny",
				"schemaVersionId": "bafkreibwyhaiseplil6tayn7spazp3qmc7nkoxdjb7uoe5zvcac4pgbwhy",
			},
		},
	}

	executeTestCase(t, test)
}
