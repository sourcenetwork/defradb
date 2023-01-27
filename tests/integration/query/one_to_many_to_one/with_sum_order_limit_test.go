// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeDescDirections(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, desc. directions.",
		Request: `query {
		    Author {
				name
			    s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
 				NewestPublishersBook: book(order: {publisher: {yearOpened: DESC}}, limit: 2) {
 					name
 				}
			}
		}`,

		Docs: map[int][]string{
			// Authors
			0: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 book
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
				// Has written no book
				`{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},

			// Books
			1: {
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 publisher
				`{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
				// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 publisher
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no publisher.
				`{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 publisher
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 publisher
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 1 Publisher
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},

			// Publishers
			2: {
				`{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
				`{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
			    }`,
				`{
					"name": "Only Publisher of Painted House",
					"address": "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
			    }`,
				`{
					"name": "Only Publisher of A Time for Mercy",
					"address": "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
			    }`,
				`{
					"name": "Only Publisher of Sooley",
					"address": "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
				"s1":   4.8 + 4.5, // Because in descending order years for John are [2020, 2013].
				"NewestPublishersBook": []map[string]any{
					{
						"name": "Theif Lord",
					},
					{
						"name": "A Time for Mercy",
					},
				},
			},
			{
				"name":                 "Not a Writer",
				"s1":                   0.0,
				"NewestPublishersBook": []map[string]any{},
			},
			{
				"name": "Cornelia Funke",
				"s1":   4.0,
				"NewestPublishersBook": []map[string]any{
					{
						"name": "The Rooster Bar",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// TODO: Fix this panic in #833.
func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeAscDirections(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, asc. directions.",
		Request: `query {
 		    Author {
 				name
 			    s1: _sum(book: {field: rating, order: {publisher: {yearOpened: ASC}}, limit: 2})
  				NewestPublishersBook: book(order: {publisher: {yearOpened: ASC}}, limit: 2) {
  					name
  				}
 			}
 		}`,

		Docs: map[int][]string{
			// Authors
			0: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
				`{
 					"name": "John Grisham",
 					"age": 65,
 					"verified": true
 				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 book
				`{
 					"name": "Cornelia Funke",
 					"age": 62,
 					"verified": false
 				}`,
				// Has written no book
				`{
 					"name": "Not a Writer",
 					"age": 6,
 					"verified": false
 				}`,
			},

			// Books
			1: {
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 publisher
				`{
 					"name": "The Rooster Bar",
 					"rating": 4,
 					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
 				}`,
				// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 publisher
				`{
 					"name": "Theif Lord",
 					"rating": 4.8,
 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
 				}`,
				// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no publisher.
				`{
 					"name": "The Associate",
 					"rating": 4.2,
 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
 				}`,
				// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 publisher
				`{
 					"name": "Painted House",
 					"rating": 4.9,
 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
 				}`,
				// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 publisher
				`{
 					"name": "A Time for Mercy",
 					"rating": 4.5,
 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
 				}`,
				// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 1 Publisher
				`{
 					"name": "Sooley",
 					"rating": 3.2,
 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
 				}`,
			},

			// Publishers
			2: {
				`{
 					"name": "Only Publisher of The Rooster Bar",
 					"address": "1 Rooster Ave., Waterloo, Ontario",
 					"yearOpened": 2022,
 					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
 			    }`,
				`{
 					"name": "Only Publisher of Theif Lord",
 					"address": "1 Theif Lord, Waterloo, Ontario",
 					"yearOpened": 2020,
 					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
 			    }`,
				`{
 					"name": "Only Publisher of Painted House",
 					"address": "600 Madison Ave., New York, New York",
 					"yearOpened": 1995,
 					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
 			    }`,
				`{
 					"name": "Only Publisher of A Time for Mercy",
 					"address": "123 Andrew Street, Flin Flon, Manitoba",
 					"yearOpened": 2013,
 					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
 			    }`,
				`{
 					"name": "Only Publisher of Sooley",
 					"address": "11 Sooley Ave., Waterloo, Ontario",
 					"yearOpened": 1999,
 					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
 			    }`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
				"s1":   4.9 + 3.2, // Because in ascending order years for John are [1995, 1999].
				"NewestPublishersBook": []map[string]any{
					{
						"name": "Painted House",
					},
					{
						"name": "Sooley",
					},
				},
			},
			{
				"name":                 "Not a Writer",
				"s1":                   0.0,
				"NewestPublishersBook": []map[string]any{},
			},
			{
				"name": "Cornelia Funke",
				"s1":   4.0,
				"NewestPublishersBook": []map[string]any{
					{
						"name": "The Rooster Bar",
					},
				},
			},
		},
	}

	testUtils.AssertPanicAndSkipChangeDetection(t, func() { executeTestCase(t, test) })
}

// TODO: Fix this panic in #833.
func TestOneToManyToOneWithSumOfDeepOrderBySubTypeOfBothDescAndAsc(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N-1 sums of deep orderby subtypes of both descending and ascending.",
		Request: `query {
		    Author {
				name
			    s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
			    s2: _sum(book: {field: rating, order: {publisher: {yearOpened: ASC}}, limit: 2})
			}
		}`,

		Docs: map[int][]string{
			// Authors
			0: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 book
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
				// Has written no book
				`{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},

			// Books
			1: {
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 publisher
				`{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
				// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 publisher
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no publisher.
				`{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 publisher
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 publisher
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 1 Publisher
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},

			// Publishers
			2: {
				`{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
				`{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
			    }`,
				`{
					"name": "Only Publisher of Painted House",
					"address": "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
			    }`,
				`{
					"name": "Only Publisher of A Time for Mercy",
					"address": "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
			    }`,
				`{
					"name": "Only Publisher of Sooley",
					"address": "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
			},
		},
		Results: []map[string]any{},
	}

	testUtils.AssertPanicAndSkipChangeDetection(t, func() { executeTestCase(t, test) })
}

// TODO: Fix this panic in #833.
func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeOppositeDirections(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, opposite directions.",
		Request: `query {
		    Author {
				name
			    s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
 				NewestPublishersBook: book(order: {publisher: {yearOpened: ASC}}, limit: 2) {
 					name
 				}
			}
		}`,

		Docs: map[int][]string{
			// Authors
			0: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 book
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
				// Has written no book
				`{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},

			// Books
			1: {
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 publisher
				`{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
				// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 publisher
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no publisher.
				`{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 publisher
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 publisher
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 1 Publisher
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},

			// Publishers
			2: {
				`{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
				`{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
			    }`,
				`{
					"name": "Only Publisher of Painted House",
					"address": "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
			    }`,
				`{
					"name": "Only Publisher of A Time for Mercy",
					"address": "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
			    }`,
				`{
					"name": "Only Publisher of Sooley",
					"address": "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
			},
		},
		Results: []map[string]any{},
	}

	testUtils.AssertPanicAndSkipChangeDetection(t, func() { executeTestCase(t, test) })
}
