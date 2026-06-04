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

package commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDAndLinkCount(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							COUNT(field: links)
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": uniqueCid, "COUNT": 0},
						{"cid": uniqueCid, "COUNT": 0},
						{"cid": uniqueCid, "COUNT": 2},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithDocUpdatesAndLinkHeadCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							fieldName
							linkCount: COUNT(field: links)
							headCount: COUNT(field: heads)
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"linkCount": 0,
							"headCount": 1,
						},
						{
							"fieldName": "_C",
							"linkCount": 1,
							"headCount": 1,
						},
						{
							"fieldName": "age",
							"linkCount": 0,
							"headCount": 0,
						},
						{
							"fieldName": "name",
							"linkCount": 0,
							"headCount": 0,
						},
						{
							"fieldName": "_C",
							"linkCount": 2,
							"headCount": 0,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
