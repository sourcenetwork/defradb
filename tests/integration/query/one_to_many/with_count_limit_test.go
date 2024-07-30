// Copyright 2022 Democratized Data Foundation
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

func TestQueryOneToManyWithCountAndLimit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count and limit",
		Request: `query {
				Author {
					name
					_count(published: {})
					published(limit: 1) {
						name
					}
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name":   "Cornelia Funke",
					"_count": 1,
					"published": []map[string]any{
						{
							"name": "Theif Lord",
						},
					},
				},
				{
					"name":   "John Grisham",
					"_count": 2,
					"published": []map[string]any{
						{
							"name": "Painted House",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountAndDifferentLimits(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count and limit",
		Request: `query {
				Author {
					name
					_count(published: {limit: 2})
					published(limit: 1) {
						name
					}
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name":   "Cornelia Funke",
					"_count": 1,
					"published": []map[string]any{
						{
							"name": "Theif Lord",
						},
					},
				},
				{
					"name":   "John Grisham",
					"_count": 2,
					"published": []map[string]any{
						{
							"name": "The Associate",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountWithLimit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count with limit",
		Request: `query {
				Author {
					name
					_count(published: {limit: 1})
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name":   "Cornelia Funke",
					"_count": 1,
				},
				{
					"name":   "John Grisham",
					"_count": 1,
				},
			},
		},
	}

	executeTestCase(t, test)
}
