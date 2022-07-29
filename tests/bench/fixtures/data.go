// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fixtures

var (
	gTypeToGQLType = map[string]string{
		"int":     "Int",
		"string":  "String",
		"float64": "Float",
		"float32": "Float",
		"bool":    "Boolean",
		"ID":      "ID",
	}
)

type User struct {
	Name     string `faker:"name"`
	Age      int
	Points   float32 `faker:"amount"`
	Verified bool
}

// #2
type Book struct {
	Name        string     `faker:"title"`
	Rating      float32    `faker:"amount"`
	Author      *Author    `fixture:"one-to-one" faker:"-"`
	Publisher   *Publisher `fixture:"one-to-many" faker:"-"`
	PublisherId ID         `faker:"-"`
}

// #3
type Author struct {
	Name     string `faker:"name"`
	Age      int
	Verified bool
	Wrote    *Book `fixture:"one-to-one,primary,0.8" faker:"-"`
	WorteId  ID
}

// #1
type Publisher struct {
	Name                 string `faker:"title"`
	PhoneNumber          string `faker:"phone_number"`
	FavouritePageNumbers []int

	// Fixture Data:
	// Rate: 50%
	// Min:1
	// Max:10
	Published []*Book `fixture:"one-to-many,0.5,1,10" faker:"-"`
}

/*

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
    favouritePageNumbers: [Int!]
    published: [book]
}

*/
