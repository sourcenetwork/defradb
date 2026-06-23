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

package branchables

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryCommitsBranchables(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.Request{
				Request: `query {
						_commits {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsBranchables_WithAllFields(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionCid := testUtils.NewSameValue()
	compositeCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	nameCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		// asserts plaintext delta bytes; encryption replaces them with ciphertext
		MultiplierExcludes: []string{multiplier.EncryptedDocs, multiplier.SignedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.Request{
				Request: `query {
						_commits {
							cid
							collectionVersionId
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}`,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":                 gomega.And(collectionCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               nil,
							"docID":               nil,
							"fieldName":           nil,
							"height":              int64(1),
							"links": []map[string]any{
								{
									"cid":       compositeCid,
									"fieldName": "_C",
								},
							},
							"heads": []map[string]any{},
						},
						{
							"cid":                 gomega.And(ageCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               testUtils.CBORValue(21),
							"docID":               "{{.DocID0_0}}",
							"fieldName":           "age",
							"height":              int64(1),
							"links":               []map[string]any{},
							"heads":               []map[string]any{},
						},
						{
							"cid":                 gomega.And(nameCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               testUtils.CBORValue("John"),
							"docID":               "{{.DocID0_0}}",
							"fieldName":           "name",
							"height":              int64(1),
							"links":               []map[string]any{},
							"heads":               []map[string]any{},
						},
						{
							"cid":                 gomega.And(compositeCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               nil,
							"docID":               "{{.DocID0_0}}",
							"fieldName":           "_C",
							"height":              int64(1),
							"links": []map[string]any{
								{
									"cid":       ageCid,
									"fieldName": "age",
								},
								{
									"cid":       nameCid,
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
