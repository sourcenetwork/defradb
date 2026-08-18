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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// A Lens module argument above 2^53 used to be rounded to the nearest representable
// float64 by every JSON decode boundary that reads an untyped Lens config into
// map[string]any without json.Decoder.UseNumber().
//
// 9007199254740993 (2^53 + 1) and 9007199254740992 (2^53) are different int64
// values, but under that bug both round to the identical float64(9007199254740992), so
// the two AddLens calls below would produce identical, content-addressed configs and get
// deduplicated into a single stored lens instead of two. Therefore, this exists as a
// regression test by checking that the count of lenses is actually two.

func TestCollectionVersionPatch_LargeIntegerLensArguments_AreNotDeduplicated(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "verified",
								"value": int64(9007199254740993), // 2^53 + 1
							},
						},
					},
				},
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "verified",
								"value": int64(9007199254740992), // 2^53
							},
						},
					},
				},
			},
			&action.ListLenses{
				ExpectedCount: immutable.Some(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
