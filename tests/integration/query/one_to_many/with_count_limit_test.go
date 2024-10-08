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
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count and limit",
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {})
						published(limit: 1) {
							name
						}
					}
				}`,
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountAndDifferentLimits(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count and limit",
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {limit: 2})
						published(limit: 1) {
							name
						}
					}
				}`,
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
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountWithLimit(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count with limit",
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {limit: 1})
					}
				}`,
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
			},
		},
	}

	executeTestCase(t, test)
}
