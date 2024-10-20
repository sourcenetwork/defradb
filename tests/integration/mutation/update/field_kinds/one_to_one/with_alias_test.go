// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromPrimarySide(t *testing.T) {
	bookID := "bae-dafb74e9-2bf1-5f12-aea9-967814592bad"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from single side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
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
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide_CollectionApi(t *testing.T) {
	author2ID := "bae-c058cfd4-259f-5b08-975d-106f13a143d5"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from secondary side",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
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
		Description: "One to one update mutation using alias relation id from secondary side",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
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
		Description: "One to one update mutation using invalid alias relation id",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
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
