// Copyright 2020 Source Inc.
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

	"github.com/sourcenetwork/defradb/db/base"

	"github.com/stretchr/testify/assert"
)

func TestSimpleOneToOneFromSingle(t *testing.T) {
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
	rm.RegisterSingle(relName1, "Author", "author", base.Meta_Relation_ONE)

	relName2, err := genRelationName("Author", "Book")
	assert.NoError(t, err)
	assert.Equal(t, relName1, relName2)
	rm.RegisterSingle(relName2, "Book", "published", base.Meta_Relation_ONE)

	_, valid := rm.validate()
	assert.True(t, valid)

	rel, err := rm.GetRelation(relName1)
	assert.NoError(t, err)
	assert.Equal(t, rel.relType, base.Meta_Relation_ONEONE)
	assert.True(t, rel.SchemaTypeIsPrimary("Author"))
	assert.False(t, rel.SchemaTypeIsPrimary("Book"))
}

func TestSimpleOneToOnePrimaryFromSingle(t *testing.T) {
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
	rm.RegisterSingle(relName1, "Author", "author", base.Meta_Relation_ONE)

	relName2, err := genRelationName("Author", "Book")
	assert.NoError(t, err)
	assert.Equal(t, relName1, relName2)
	rm.RegisterSingle(relName2, "Book", "published", base.Meta_Relation_ONE|base.Meta_Relation_Primary)

	_, valid := rm.validate()
	assert.True(t, valid)

	rel, err := rm.GetRelation(relName1)
	assert.NoError(t, err)
	assert.Equal(t, rel.relType, base.Meta_Relation_ONEONE)
	assert.True(t, rel.SchemaTypeIsPrimary("Book"))
	assert.False(t, rel.SchemaTypeIsPrimary("Author"))
}
