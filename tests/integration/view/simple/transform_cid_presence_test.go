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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// AddViewOptions.TransformCID being omitted (None) and being explicitly set to ""
// (Some("")) must not be treated the same. An empty CID is invalid and should be
// rejected, while an omitted one means "no transform". The latter is alreedy covered
// by existing tests, so we only test the former here.

func TestAddView_BlankTransformCID_FailsAndCreatesNoView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
				TransformCID:  immutable.Some(""),
				ExpectedError: "cid too short",
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("UserView"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
