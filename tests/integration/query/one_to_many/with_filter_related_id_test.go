// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryFromManySideWithEqFilterOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query from many side with _eq filter on related field type.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {author: {_docID: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "The Client"},
						{"name": "A Time for Mercy"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithFilterOnRelatedObjectID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query from many side with filter on related field.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(filter: {author_id: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "The Client"},
						{"name": "A Time for Mercy"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithSameFiltersInDifferentWayOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query from many side with same filters in different way on related type.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(
						filter: {
							author: {_docID: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}},
							author_id: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "The Client"},
						{"name": "A Time for Mercy"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromSingleSideWithEqFilterOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query from single side with _eq filter on related field type.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {published: {_docID: {_eq: "bae-064f13c1-7726-5d53-8eec-c395d94da4d0"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromSingleSideWithFilterOnRelatedObjectID_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query from single side with filter on related field.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {published_id: {_eq: "bae-064f13c1-7726-5d53-8eec-c395d94da4d0"}}) {
						name
					}
				}`,
				ExpectedError: "Argument \"filter\" has invalid value {published_id: {_eq: \"bae-064f13c1-7726-5d53-8eec-c395d94da4d0\"}}.\nIn field \"published_id\": Unknown field.",
			},
		},
	}

	executeTestCase(t, test)
}
