package db

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document"
	"github.com/stretchr/testify/assert"
)

type queryTestCase struct {
	description string
	query       string
	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	docs map[int][]string
	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	updates map[int][]string
	results []map[string]interface{}
}

func TestQuerySimple(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple query with no filter",
			query: `query {
						users {
							_key
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"_key": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with alias, no filter",
			query: `query {
						users {
							username: Name
							age: Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"username": "John",
					"age":      uint64(21),
				},
			},
		},
		{
			description: "Simple query with no filter, mutiple rows",
			query: `query {
						users {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 27
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(27),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter (key)",
			query: `query {
						users(filter: {_key: {_eq: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter (Name)",
			query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter and selection",
			query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Name					
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
				},
			},
		},
		{
			description: "Simple query with basic filter and selection (diff from filter)",
			query: `query {
						users(filter: {Name: {_eq: "John"}}) {
							Age					
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Age": uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(name), no results",
			query: `query {
						users(filter: {Name: {_eq: "Bob"}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_eq: 21}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(age), greater than",
			query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_gt: 40}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			results: []map[string]interface{}{},
		},
		{
			description: "Simple query with basic filter(age)",
			query: `query {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic limit",
			query: `query {
						users(limit: 1) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				// {
				// 	"Name": "John",
				// 	"Age":  uint64(21),
				// },
			},
		},
		{
			description: "Simple query with basic limit & offset",
			query: `query {
						users(limit: 1, offset: 1) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			results: []map[string]interface{}{
				// {
				// 	"Name": "Bob",
				// 	"Age":  uint64(32),
				// },
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with basic limit, more rows",
			query: `query {
						users(limit: 2) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
			},
		},
		{
			description: "Simple query with basic limit & offset, more rows",
			query: `query {
						users(limit: 2, offset: 2) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
			},
		},
		{
			description: "Simple query with basic sort ASC",
			query: `query {
						users(order: {Age: ASC}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
			},
		},
		{
			description: "Simple query with basic sort DESC",
			query: `query {
						users(order: {Age: DESC}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
			},
		},
		{
			description: "Simple query with compound sort",
			query: `query {
						users(order: {Age: DESC, Verified: ASC}) {
							Name
							Age
							Verified
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21,
					"Verified": true
				}`),
					(`{
					"Name": "Bob",
					"Age": 21,
					"Verified": false
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`),
					(`{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name":     "Carlo",
					"Age":      uint64(55),
					"Verified": true,
				},
				{
					"Name":     "Bob",
					"Age":      uint64(21),
					"Verified": false,
				},
				{
					"Name":     "John",
					"Age":      uint64(21),
					"Verified": true,
				},
				{
					"Name":     "Alice",
					"Age":      uint64(19),
					"Verified": false,
				},
			},
		},
		{
			description: "Simple query with sort & filter",
			query: `query {
						users(filter: {Age: {_gt: 30}}, order: {Age: DESC}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
			},
		},
		{
			description: "Simple query with logical compound filter (and)",
			query: `query {
						users(filter: {_and: [{Age: {_gt: 20}}, {Age: {_lt: 50}}]}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			description: "Simple query with logical compound filter (or)",
			query: `query {
						users(filter: {_or: [{Age: {_eq: 55}}, {Age: {_eq: 19}}]}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
			},
		},
		{
			description: "Simple query with special filter (or)",
			query: `query {
						users(filter: {Age: {_in: [19, 40, 55]}}) {
							Name
							Age
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryRelationOne(t *testing.T) {
	var bookAuthorGQLSchema = (`
	type book {
		name: String
		rating: Float
		author: author 
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		published: book @primary
	}
	`)

	tests := []queryTestCase{
		{
			description: "One-to-one relation query with no filter",
			query: `query {
						book {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9
				}`)},
				//authors
				1: []string{ // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`)},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			description: "One-to-one relation query with simple filter on sub type",
			query: `query {
						book {
							name
							rating
							author(filter: {age: {_eq: 65}}) {
								name
								age
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9
				}`)},
				//authors
				1: []string{ // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`)},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			description: "One-to-one relation query with simple filter on parent",
			query: `query {
						book(filter: {name: {_eq: "Painted House"}}) {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9
				}`)},
				//authors
				1: []string{ // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`)},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			description: "One-to-one relation query with simple sub filter on parent",
			query: `query {
						book(filter: {author: {verified: {_eq: true}}}) {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9
				}`)},
				//authors
				1: []string{ // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`)},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			description: "One-to-one relation query with simple sort by sub type",
			query: `query {
				book(order: {author: {verified: DESC}}) {
					name
					rating
					author {
						name
						age
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{
					// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9
					}`),
					// bae-d432bdfb-787d-5a1c-ac29-dc025ab80095
					(`{
					"name": "Theif Lord",
					"rating": 4.8
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"published_id": "bae-d432bdfb-787d-5a1c-ac29-dc025ab80095"
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
				{
					"name":   "Theif Lord",
					"rating": 4.8,
					"author": map[string]interface{}{
						"name": "Cornelia Funke",
						"age":  uint64(62),
					},
				},
			},
		},
		{
			description: "One-to-one relation secondary direction, no filter",
			query: `query {
						author {
							name
							age
							published {
								name
								rating
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{
					// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
							"name": "Painted House",
							"rating": 4.9
							}`),
					// bae-d432bdfb-787d-5a1c-ac29-dc025ab80095
					// (`{
					// "name": "Theif Lord",
					// "rating": 4.8
					// }`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
							"name": "John Grisham",
							"age": 65,
							"verified": true,
							"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
							}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					// (`{
					// "name": "Cornelia Funke",
					// "age": 62,
					// "verified": false,
					// "published_id": "bae-d432bdfb-787d-5a1c-ac29-dc025ab80095"
					// }`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": map[string]interface{}{
						"name":   "Painted House",
						"rating": 4.9,
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(bookAuthorGQLSchema)
		assert.NoError(t, err)

		// bookDesc := newTestQueryCollectionDescription2()
		bookCol, err := db.GetCollection("book")
		assert.NoError(t, err)

		// authorDesc := newTestQueryCollectionDescription3()
		authorCol, err := db.GetCollection("author")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{bookCol, authorCol}, test)
	}

}

func TestQueryRelationMany(t *testing.T) {
	var bookAuthorGQLSchema = (`
	type book {
		name: String
		rating: Float
		author: author
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		published: [book]
	}
	`)

	tests := []queryTestCase{
		{
			description: "One-to-many relation query from one side",
			query: `query {
						book {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`)},
				//authors
				1: []string{ // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`)},
			},
			results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			description: "One-to-many relation query from many side",
			query: `query {
				author {
					name
					age
					published {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
				},
				{
					"name": "Cornelia Funke",
					"age":  uint64(62),
					"published": []map[string]interface{}{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
				},
			},
		},
		{
			description: "One-to-many relation query from the many side, simple filter",
			query: `query {
				author(filter: {age: {_gt: 63}}) {
					name
					age
					published {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
				},
			},
		},
		{
			description: "One-to-many relation query from the many side, simple filter on root and sub type",
			query: `query {
				author(filter: {age: {_gt: 63}}) {
					name
					age
					published(filter: {rating: {_gt: 4.6}}) {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
		{
			description: "One-to-many relation query from the many side, order on sub",
			query: `query {
				author(filter: {age: {_gt: 63}}) {
					name
					age
					published(order: {rating: ASC}) {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
		{
			description: "One-to-many relation query from the many side, order & limit on sub",
			query: `query {
				author(filter: {age: {_gt: 63}}) {
					name
					age
					published(order: {rating: ASC}, limit: 1) {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{ 
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
				},
			},
		},
		{
			description: "One-to-many relation query from the many side, filter on sub from root",
			query: `query {
				author(filter: {published: {rating: {_gt: 4.1}}}) {
					name
					age
					published(order: {rating: DESC}) {
						name
						rating
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{ // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					(`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
					(`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`),
					(`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				},
				//authors
				1: []string{
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					(`{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`),
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					(`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`),
				},
			},
			results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": []map[string]interface{}{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
				},
				{
					"name": "Cornelia Funke",
					"age":  uint64(62),
					"published": []map[string]interface{}{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(bookAuthorGQLSchema)
		assert.NoError(t, err)

		// bookDesc := newTestQueryCollectionDescription2()
		bookCol, err := db.GetCollection("book")
		assert.NoError(t, err)

		// authorDesc := newTestQueryCollectionDescription3()
		authorCol, err := db.GetCollection("author")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{bookCol, authorCol}, test)
	}
}

func runQueryTestCase(t *testing.T, db *DB, collections []client.Collection, test queryTestCase) {
	// insert docs
	for cid, docs := range test.docs {
		for i, docStr := range docs {
			doc, err := document.NewFromJSON([]byte(docStr))
			assert.NoError(t, err, test.description)
			err = collections[cid].Save(doc)
			assert.NoError(t, err, test.description)

			// check for updates
			updates, ok := test.updates[i]
			if ok {
				for _, u := range updates {
					err = doc.SetWithJSON([]byte(u))
					assert.NoError(t, err, test.description)
					err = collections[cid].Save(doc)
					assert.NoError(t, err, test.description)
				}
			}
		}
	}

	// exec query
	txn, err := db.NewTxn(true)
	assert.NoError(t, err, test.description)
	results, err := db.queryExecutor.ExecQuery(db, txn, test.query)
	assert.NoError(t, err, test.description)

	fmt.Println(test.description)
	fmt.Println(results)
	fmt.Println("--------------")
	fmt.Println("")

	// compare results
	assert.Equal(t, len(test.results), len(results), test.description)
	for i, result := range results {
		assert.Equal(t, test.results[i], result, test.description)
	}
}

func TestMutationCreateSimple(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
			_key
			name
			age
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, map[string]interface{}{
		"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
		"age":  int64(27),
		"name": "John",
	}, results[0])

}

func TestMutationUpdateFilterSimple(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		update_user(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
			_key
			name
			points
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	col, err := db.GetCollection("user")
	assert.NoError(t, err)

	doc1, err := document.NewFromJSON([]byte(`{
		"name": "John",
		"age": 27,
		"verified": true,
		"points": 42.1
	}`))
	assert.NoError(t, err)

	err = col.Save(doc1)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, map[string]interface{}{
		"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
		"name":   "John",
		"points": float64(59),
	}, results[0])
}

func TestMutationUpdateFilterMultiDocsSingleResult(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		update_user(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
			_key
			name
			points
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	col, err := db.GetCollection("user")
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"name": "John",
		"age": 27,
		"verified": true,
		"points": 42.1
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	doc, err = document.NewFromJSON([]byte(`{
		"name": "Bob",
		"age": 39,
		"verified": false,
		"points": 66.6
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, map[string]interface{}{
		"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
		"name":   "John",
		"points": float64(59),
	}, results[0])
}

func TestMutationUpdateFilterMultiDocsMultiResult(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		update_user(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
			_key
			name
			points
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	col, err := db.GetCollection("user")
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"name": "John",
		"age": 27,
		"verified": true,
		"points": 42.1
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	doc, err = document.NewFromJSON([]byte(`{
		"name": "Bob",
		"age": 39,
		"verified": true,
		"points": 66.6
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 2)
	assert.Equal(t, []map[string]interface{}{
		{
			"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
			"name":   "John",
			"points": float64(59),
		},
		{
			"_key":   "bae-455b5896-6203-582f-b46e-729c53a2d14b",
			"name":   "Bob",
			"points": float64(59),
		},
	}, results)
}

func TestMutationUpdateByKeyMultiDocsSingleResult(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		update_user(id: "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d", data: "{\"points\": 59}") {
			_key
			name
			points
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	col, err := db.GetCollection("user")
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"name": "John",
		"age": 27,
		"verified": true,
		"points": 42.1
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	doc, err = document.NewFromJSON([]byte(`{
		"name": "Bob",
		"age": 39,
		"verified": false,
		"points": 66.6
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, map[string]interface{}{
		"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
		"name":   "John",
		"points": float64(59),
	}, results[0])
}

func TestMutationUpdateByKeysMultiDocsMultiResult(t *testing.T) {

	userSchema := `
	type user {
		name: String
		age: Int
		points: Float
		verified: Boolean
	}
	`

	// data :=

	query := `
	mutation {
		update_user(ids: ["bae-0a24cf29-b2c2-5861-9d00-abd6250c475d", "bae-958c9334-73cf-5695-bf06-cf06826babfa"], data: "{\"points\": 59}") {
			_key
			name
			points
		}
	}`

	db, err := newMemoryDB()
	assert.NoError(t, err)

	err = db.LoadSchema(userSchema)
	assert.NoError(t, err)

	col, err := db.GetCollection("user")
	assert.NoError(t, err)

	doc, err := document.NewFromJSON([]byte(`{
		"name": "John",
		"age": 27,
		"verified": true,
		"points": 42.1
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	doc, err = document.NewFromJSON([]byte(`{
		"name": "Bob",
		"age": 39,
		"verified": false,
		"points": 66.6
	}`))
	assert.NoError(t, err)
	err = col.Save(doc)
	assert.NoError(t, err)

	// exec query
	txn, err := db.NewTxn(false)
	assert.NoError(t, err)
	results, err := db.queryExecutor.ExecQuery(db, txn, query)
	assert.NoError(t, err)

	assert.Len(t, results, 2)
	assert.Equal(t, []map[string]interface{}{
		{
			"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
			"name":   "John",
			"points": float64(59),
		},
		{
			"_key":   "bae-958c9334-73cf-5695-bf06-cf06826babfa",
			"name":   "Bob",
			"points": float64(59),
		},
	}, results)
}

func TestQueryMultiNodeSelectionOne(t *testing.T) {
	var bookAuthorPublisherGQLSchema = (`
	type book {
		name: String
		rating: Float
		author: author 
		publisher: publisher
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		wrote: book @primary
	}

	type publisher {
		name: String
		address: String
		published: book
	}
	`)

	tests := []queryTestCase{
		{
			description: "multinode: One-to-one relation query with no filter",
			query: `query {
				book {
					name
					author {
						name
					}
					publisher {
						name
					}
				}
			}`,
			docs: map[int][]string{
				//books
				0: []string{
					// bae-7e5ae688-3a77-5b4f-a74c-59301bd1eb25
					(`{
						"name": "The Coffee Table Book",
						"rating": 4.9,
						"publisher_id": "bae-81804a20-4d08-509e-a3e8-fd770622a356"
					}`)},
				//authors
				1: []string{
					// bae-5eae6a8a-0c52-535c-9c20-df42b7044e20
					(`{
						"name": "Cosmo Kramer",
						"age": 44,
						"verified": true,
						"wrote_id": "bae-7e5ae688-3a77-5b4f-a74c-59301bd1eb25"
					}`)},
				// publishers
				2: []string{
					// bae-81804a20-4d08-509e-a3e8-fd770622a356
					(`{
						"name": "Pendant Publishing",
						"address": "600 Madison Ave., New York, New York"
					}`)},
			},
			results: []map[string]interface{}{
				{
					"name": "The Coffee Table Book",
					"author": map[string]interface{}{
						"name": "Cosmo Kramer",
					},
					"publisher": map[string]interface{}{
						"name": "Pendant Publishing",
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(bookAuthorPublisherGQLSchema)
		assert.NoError(t, err)

		// bookDesc := newTestQueryCollectionDescription2()
		bookCol, err := db.GetCollection("book")
		assert.NoError(t, err)

		// authorDesc := newTestQueryCollectionDescription3()
		authorCol, err := db.GetCollection("author")
		assert.NoError(t, err)

		// authorDesc := newTestQueryCollectionDescription3()
		pubCol, err := db.GetCollection("publisher")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{bookCol, authorCol, pubCol}, test)
	}
}

func TestQueryLatestCommits(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple latest commits query",
			query: `query {
						latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"cid": "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"links": []map[string]interface{}{
						{
							"cid":  "QmPaY2DNmd7LtRDpReswc5UTGoU5Q32Py1aEVG7Shq6Np1",
							"name": "Age",
						},
						{
							"cid":  "Qmag2zKKGGQwVSss9pQn3hjTu9opdF5mkJXUR9rt2A651h",
							"name": "Name",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryAllCommitsSingleDAG(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple latest commits query",
			query: `query {
						allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"cid": "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"links": []map[string]interface{}{
						{
							"cid":  "QmPaY2DNmd7LtRDpReswc5UTGoU5Q32Py1aEVG7Shq6Np1",
							"name": "Age",
						},
						{
							"cid":  "Qmag2zKKGGQwVSss9pQn3hjTu9opdF5mkJXUR9rt2A651h",
							"name": "Name",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryAllCommitsMultipleDAG(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Simple latest commits query",
			query: `query {
						allCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							height
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			updates: map[int][]string{
				0: []string{
					(`{"Age": 22}`), // update to change age to 22 on document 0
				},
			},
			results: []map[string]interface{}{
				{
					"cid":    "QmQQgYgC3PLFCTwsSgMHHFvFbPEeWDKkbsnvYJwuLP3R8t",
					"height": int64(2),
				},
				{
					"cid":    "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"height": int64(1),
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryEmbeddedLatestCommit(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "Embedded latest commits query within object query",
			query: `query {
						users {
							Name
							Age
							_version {
								cid
								links {
									cid
									name
								}
							}
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
					"_version": []map[string]interface{}{
						{
							"cid": "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
							"links": []map[string]interface{}{
								{
									"cid":  "QmPaY2DNmd7LtRDpReswc5UTGoU5Q32Py1aEVG7Shq6Np1",
									"name": "Age",
								},
								{
									"cid":  "Qmag2zKKGGQwVSss9pQn3hjTu9opdF5mkJXUR9rt2A651h",
									"name": "Name",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

func TestQueryOneCommit(t *testing.T) {
	var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
	`)

	tests := []queryTestCase{
		{
			description: "query for a single block by CID",
			query: `query {
						commit(cid: "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV") {
							cid
							height
							delta
						}
					}`,
			docs: map[int][]string{
				0: []string{
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			results: []map[string]interface{}{
				{
					"cid":    "QmaXdKKsc5GRWXtMytZj4PEf5hFgFxjZaKToQpDY8cAocV",
					"height": int64(1),
					// cbor encoded delta
					"delta": []uint8{0xa2, 0x63, 0x41, 0x67, 0x65, 0x15, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x64, 0x4a, 0x6f, 0x68, 0x6e},
				},
			},
		},
	}

	for _, test := range tests {
		db, err := newMemoryDB()
		assert.NoError(t, err)

		err = db.LoadSchema(userCollectionGQLSchema)
		assert.NoError(t, err)

		// desc := newTestQueryCollectionDescription1()
		col, err := db.GetCollection("users")
		assert.NoError(t, err)

		runQueryTestCase(t, db, []client.Collection{col}, test)
	}

}

// var userCollectionGQLSchema = (`
// type users {
// 	Name: String
// 	Age: Int
// }
// `)

// // func newQueryableDB()

// func TestSimpleCollectionQuery(t *testing.T) {
// 	db, err := newMemoryDB()
// 	assert.NoError(t, err)

// 	desc := newTestCollectionDescription()
// 	col, err := db.CreateCollection(desc)
// 	assert.NoError(t, err)

// 	executor, err := planner.NewQueryExecutor()
// 	assert.NoError(t, err)

// 	err = executor.Generator.FromSDL(userCollectionGQLSchema)
// 	assert.NoError(t, err)

// 	doc1, err := document.NewFromJSON([]byte(`{
// 		"Name": "John",
// 		"Age": 21
// 	}`))

// 	assert.NoError(t, err)
// 	err = col.Save(doc1)
// 	assert.NoError(t, err)

// 	txn, err := db.NewTxn(true)
// 	assert.NoError(t, err)

// 	// obj := executor.SchemaManager.Schema().TypeMap()["users"].(*gql.Object)
// 	// obj.Fields()
// 	// spew.Dump(obj.Fields())

// 	var userQuery = (`
// 	query {
// 		users {
// 			Name
// 			Age
// 		}
// 	}`)

// 	docs, err := executor.ExecQuery(txn, userQuery)
// 	assert.NoError(t, err)

// 	fmt.Println(docs)
// 	assert.True(t, len(docs) == 1)
// }

// func TestSimpleCollectionQueryWithFilter(t *testing.T) {
// 	db, err := newMemoryDB()
// 	assert.NoError(t, err)

// 	desc := newTestCollectionDescription()
// 	col, err := db.CreateCollection(desc)
// 	assert.NoError(t, err)

// 	executor, err := planner.NewQueryExecutor()
// 	assert.NoError(t, err)

// 	err = executor.Generator.FromSDL(userCollectionGQLSchema)
// 	assert.NoError(t, err)

// 	doc1, err := document.NewFromJSON([]byte(`{
// 		"Name": "John",
// 		"Age": 21
// 	}`))

// 	assert.NoError(t, err)
// 	err = col.Save(doc1)
// 	assert.NoError(t, err)

// 	txn, err := db.NewTxn(true)
// 	assert.NoError(t, err)

// 	// obj := executor.SchemaManager.Schema().TypeMap()["users"].(*gql.Object)
// 	// obj.Fields()
// 	// spew.Dump(obj.Fields())

// 	var userQuery = (`
// 	query {
// 		users(filter: {Name: {_eq: "John"}}) {
// 			Name
// 			Age
// 		}
// 	}`)

// 	docs, err := executor.ExecQuery(txn, userQuery)
// 	assert.NoError(t, err)

// 	// fmt.Println(docs)
// 	assert.Len(t, docs, 1)

// 	assert.Equal(t, map[string]interface{}{
// 		"Name": "John",
// 		"Age":  uint64(21),
// 	}, docs[0])
// }

// func newTestQueryCollectionDescription1() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "users",
// 		ID:   uint32(1),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(1),
// 			FieldIDs: []uint32{1, 2, 3, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "Name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "Age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "Verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription2() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "book",
// 		ID:   uint32(2),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(2),
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "rating",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_FLOAT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "author",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "author",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEONE | base.Meta_Relation_Primary,
// 				},
// 				base.FieldDescription{
// 					Name: "author_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription3() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "author",
// 		ID:   uint32(3),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(3),
// 			Name:     "author",
// 			FieldIDs: []uint32{1, 2, 3, 4, 5, 6},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "published",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "book",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE,
// 				},
// 				base.FieldDescription{
// 					Name: "published_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription4() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "book",
// 		ID:   uint32(2),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(2),
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "rating",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_FLOAT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "author",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT,
// 					Schema: "author",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_ONE | base.Meta_Relation_ONEMANY | base.Meta_Relation_Primary,
// 				},
// 				base.FieldDescription{
// 					Name: "author_id",
// 					ID:   base.FieldID(6),
// 					Kind: base.FieldKind_DocKey,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }

// func newTestQueryCollectionDescription5() base.CollectionDescription {
// 	return base.CollectionDescription{
// 		Name: "author",
// 		ID:   uint32(3),
// 		Schema: base.SchemaDescription{
// 			ID:       uint32(3),
// 			Name:     "author",
// 			FieldIDs: []uint32{1, 2, 3, 4, 5},
// 			Fields: []base.FieldDescription{
// 				base.FieldDescription{
// 					Name: "_key",
// 					ID:   base.FieldID(1),
// 					Kind: base.FieldKind_DocKey,
// 				},
// 				base.FieldDescription{
// 					Name: "name",
// 					ID:   base.FieldID(2),
// 					Kind: base.FieldKind_STRING,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "age",
// 					ID:   base.FieldID(3),
// 					Kind: base.FieldKind_INT,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name: "verified",
// 					ID:   base.FieldID(4),
// 					Kind: base.FieldKind_BOOL,
// 					Typ:  core.LWW_REGISTER,
// 				},
// 				base.FieldDescription{
// 					Name:   "published",
// 					ID:     base.FieldID(5),
// 					Kind:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
// 					Schema: "book",
// 					Typ:    core.NONE_CRDT,
// 					Meta:   base.Meta_Relation_MANY | base.Meta_Relation_ONEMANY,
// 				},
// 			},
// 		},
// 		Indexes: []base.IndexDescription{
// 			base.IndexDescription{
// 				Name:    "primary",
// 				ID:      uint32(0),
// 				Primary: true,
// 				Unique:  true,
// 			},
// 		},
// 	}
// }
