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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// Two documents with the same "Name" value share one genesis "Name" field block, so its CID is
// owned by both documents. A time-travel query against that shared CID returns every owning
// document's state at that version - not an error, and not just one arbitrary owner.
func TestQuerySimple_WithSharedFieldCid_ReturnsAllOwners(t *testing.T) {
	test := testUtils.TestCase{
		// templated CIDs would change under encryption or signing.
		MultiplierExcludes: []string{multiplier.EncryptedDocs, multiplier.SignedDocs},
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Shared",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Shared",
					"Age": 22
				}`,
			},
			&action.Request{
				Request: `query {
					Users (cid: "{{.FieldCID0_0_Name_0}}") {
						_docID
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"_docID": testUtils.NewDocIndex(0, 0), "Name": "Shared"},
						{"_docID": testUtils.NewDocIndex(0, 1), "Name": "Shared"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
