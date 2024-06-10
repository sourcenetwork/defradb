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
	test := testUtils.RequestTestCase{

		Description: "One-to-many query from many side with _eq filter on related field type.",

		Request: `query {
			Book(filter: {author: {_docID: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}}}) {
				name
			}
		}`,

		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{"name": "The Client"},
			{"name": "Painted House"},
			{"name": "A Time for Mercy"},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithFilterOnRelatedObjectID(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "One-to-many query from many side with filter on related field.",

		Request: `query {
			Book(filter: {author_id: {_eq: "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"}}) {
				name
			}
		}`,

		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{"name": "The Client"},
			{"name": "Painted House"},
			{"name": "A Time for Mercy"},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithSameFiltersInDifferentWayOnRelatedType(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "One-to-many query from many side with same filters in different way on related type.",

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

		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{"name": "The Client"},
			{"name": "Painted House"},
			{"name": "A Time for Mercy"},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromSingleSideWithEqFilterOnRelatedType(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "One-to-many query from single side with _eq filter on related field type.",

		Request: `query {
			Author(filter: {published: {_docID: {_eq: "bae-96c9de0f-2903-5589-9604-b42882afde8c"}}}) {
				name
			}
		}`,

		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromSingleSideWithFilterOnRelatedObjectID_Error(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "One-to-many query from single side with filter on related field.",

		Request: `query {
			Author(filter: {published_id: {_eq: "bae-5366ba09-54e8-5381-8169-a770aa9282ae"}}) {
				name
			}
		}`,

		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},

		ExpectedError: "Argument \"filter\" has invalid value {published_id: {_eq: \"bae-5366ba09-54e8-5381-8169-a770aa9282ae\"}}.\nIn field \"published_id\": Unknown field.",
	}

	executeTestCase(t, test)
}
