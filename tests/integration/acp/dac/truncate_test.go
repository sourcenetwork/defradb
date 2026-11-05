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

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_DAC_TruncateSoftDeletedIndexedDocWithoutReadAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - name: delete
  - name: read
  - name: update
`,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{"name":"alice"}`,
			},
			testUtils.DeleteDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				DocID:        0,
			},
			&action.Truncate{
				Identity:        testUtils.ClientIdentity(2),
				CollectionIndex: 0,
				DocIndexes:      []int{0},
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc:          `{"name":"alice"}`,
			},
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users(filter: {name: {_eq: "alice"}}) { name }
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "alice"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
