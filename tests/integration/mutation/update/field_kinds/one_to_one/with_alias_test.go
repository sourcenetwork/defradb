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

package one_to_one

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromPrimarySide(t *testing.T) {
	bookID := "bae-9164d9cb-db28-5e2b-9d87-31afd65945d0"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        1,
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					bookID,
				),
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide_CollectionApi(t *testing.T) {
	author2ID := "bae-c058cfd4-259f-5b08-975d-106f13a143d5"

	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionSaveMutationType,
			state.CollectionNamedMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "cannot set relation from secondary side",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide_GQL(t *testing.T) {
	author2ID := "bae-c058cfd4-259f-5b08-975d-106f13a143d5"

	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "Argument \"input\" has invalid value",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasWithInvalidLengthRelationIDToLink_Error(t *testing.T) {
	invalidLenSubID := "35953ca-518d-9e6b-9ce6cd00eff5"
	invalidBookID := "bae-" + invalidLenSubID

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					invalidBookID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string \"" + invalidLenSubID + "\"",
			},
		},
	}

	executeTestCase(t, test)
}
