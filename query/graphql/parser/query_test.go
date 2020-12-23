package parser

import (
	"testing"

	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/assert"
)

func TestParse_Query(t *testing.T) {

}

func TestParse_Limit_Limit(t *testing.T) {
	var query = (`
	query {
		users(limit: 10)
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	limit := q.Queries[0].Selections[0].(*Select).Limit
	assert.Equal(t, limit.Limit, int64(10))
	assert.Equal(t, limit.Offset, int64(0))
}

func TestParse_Limit_Offset(t *testing.T) {
	var query = (`
	query {
		users(offset: 100)
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	limit := q.Queries[0].Selections[0].(*Select).Limit
	assert.Equal(t, limit.Limit, int64(0))
	assert.Equal(t, limit.Offset, int64(100))
}

func TestParse_Limit_Both(t *testing.T) {
	var query = (`
	query {
		users(limit: 1, offset: 100)
	}`)

	source := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "",
	})

	doc, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	assert.NoError(t, err)

	q, err := ParseQuery(doc)
	assert.NoError(t, err)

	limit := q.Queries[0].Selections[0].(*Select).Limit
	assert.Equal(t, limit.Limit, int64(1))
	assert.Equal(t, limit.Offset, int64(100))
}
