// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
)

func TestSimpleOneToOneFromSingle(t *testing.T) {
	t.Parallel()
	rm := NewRelationManager()

	/*
		type Book {
			title: String
			author: Author
		}

		type Author {
			name: String
			published: Book
		}

		// without explicit @primary directive
		// Author is auto set to primary
	*/
	relName1, err := genRelationName("Book", "Author")
	assert.NoError(t, err)
	rm.RegisterSingle(relName1, "Author", "author", client.Relation_Type_ONE)

	relName2, err := genRelationName("Author", "Book")
	assert.NoError(t, err)
	assert.Equal(t, relName1, relName2)
	rm.RegisterSingle(relName2, "Book", "published", client.Relation_Type_ONE)

	rel, err := rm.GetRelation(relName1)
	assert.NoError(t, err)
	assert.Equal(t, rel.relType, client.Relation_Type_ONEONE)
}

func TestSimpleOneToOnePrimaryFromSingle(t *testing.T) {
	t.Parallel()
	rm := NewRelationManager()

	/*
		type Book {
			title: String
			author: Author
		}

		type Author {
			name: String
			published: Book
		}

		// without explicit @primary directive
		// Author is auto set to primary
	*/
	relName1, err := genRelationName("Book", "Author")
	assert.NoError(t, err)
	rm.RegisterSingle(relName1, "Author", "author", client.Relation_Type_ONE)

	relName2, err := genRelationName("Author", "Book")
	assert.NoError(t, err)
	assert.Equal(t, relName1, relName2)
	rm.RegisterSingle(
		relName2,
		"Book",
		"published",
		client.Relation_Type_ONE|client.Relation_Type_Primary,
	)

	rel, err := rm.GetRelation(relName1)
	assert.NoError(t, err)
	assert.Equal(t, rel.relType, client.Relation_Type_ONEONE)
}
