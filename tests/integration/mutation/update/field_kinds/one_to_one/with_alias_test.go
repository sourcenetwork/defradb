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
	author1ID := "bae-53eff350-ad8e-532c-b72d-f95c4f47909c"
	bookID := "bae-89d64ba1-44e3-5d75-a610-7226077ece48"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from single side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					author1ID,
				),
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

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide(t *testing.T) {
	author1ID := "bae-53eff350-ad8e-532c-b72d-f95c4f47909c"
	author2ID := "bae-c058cfd4-259f-5b08-975d-106f13a143d5"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					author1ID,
				),
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
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasWithInvalidLengthRelationIDToLink_Error(t *testing.T) {
	author1ID := "bae-53eff350-ad8e-532c-b72d-f95c4f47909c"
	invalidLenSubID := "35953ca-518d-9e6b-9ce6cd00eff5"
	invalidAuthorID := "bae-" + invalidLenSubID

	test := testUtils.TestCase{
		Description: "One to one update mutation using invalid alias relation id",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string \"" + invalidLenSubID + "\"",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_InvalidAliasRelationNameToLinkFromSecondarySide_Error(t *testing.T) {
	author1ID := "bae-53eff350-ad8e-532c-b72d-f95c4f47909c"
	invalidAuthorID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ee"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySideWithWrongField_Error(t *testing.T) {
	author1ID := "bae-53eff350-ad8e-532c-b72d-f95c4f47909c"
	author2ID := "bae-c058cfd4-259f-5b08-975d-106f13a143d5"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation alias name from secondary side, with a wrong field.",
		// This restiction is temporary due to a bug in the collection api, see
		// https://github.com/sourcenetwork/defradb/issues/1703 for more info.
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"notName": "Unpainted Condo",
						"author": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "Unknown field.",
			},
		},
	}

	executeTestCase(t, test)
}
