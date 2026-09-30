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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// Two documents sharing the same "name" value produce a byte-identical genesis "name" field
// block, so a single CID is owned by both documents.

// A commits query by the shared field CID combined with a docID resolves to the requested
// document: the docID disambiguates which owner of the shared block is meant.
func TestQueryCommits_WithSharedFieldCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are hardcoded/matched because template placeholders are not
		// resolved inside Request.Results. See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	22
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: "{{.FieldCID0_0_name_0}}", docID: "{{.DocID0_0}}") {
							cid
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       testUtils.ValidCID(),
							"fieldName": "name",
							"docID":     "{{.DocID0_0}}",
						},
					},
				},
			},
			&action.Request{
				Request: `query {
						_commits(cid: "{{.FieldCID0_1_name_0}}", docID: "{{.DocID0_1}}") {
							cid
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       testUtils.ValidCID(),
							"fieldName": "name",
							"docID":     "{{.DocID0_1}}",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A commits query by a shared field CID without a docID returns the commit for every
// document that owns the block: the same block is part of both documents' histories, so it
// is reported once per owner rather than dropped or attributed to a single arbitrary owner.
func TestQueryCommits_WithSharedFieldCidWithoutDocID_ReturnsAllOwners(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are matched because template placeholders are not resolved inside
		// Request.Results. See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	22
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: "{{.FieldCID0_0_name_0}}") {
							cid
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       testUtils.ValidCID(),
							"fieldName": "name",
							"docID":     "{{.DocID0_0}}",
						},
						{
							"cid":       testUtils.ValidCID(),
							"fieldName": "name",
							"docID":     "{{.DocID0_1}}",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
