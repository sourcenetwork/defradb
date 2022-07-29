package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagParserOneToOnePrimary(t *testing.T) {
	fixtureTagStr := "one-to-one,primary,0.8"
	tg, err := parseTag(fixtureTagStr)
	require.NoError(t, err)

	assert.Equal(t, tag{
		rel:       oneToOne,
		isPrimary: true,
		fillRate:  0.8,
	}, tg)
}

func TestTagParserOneToOneSecondary(t *testing.T) {
	fixtureTagStr := "one-to-one"
	tg, err := parseTag(fixtureTagStr)
	require.NoError(t, err)

	assert.Equal(t, tag{
		rel:       oneToOne,
		isPrimary: false,
	}, tg)
}

func TestTagParserOneToMany(t *testing.T) {
	fixtureTagStr := "one-to-many,0.5,1,10"
	tg, err := parseTag(fixtureTagStr)
	require.NoError(t, err)

	assert.Equal(t, tag{
		rel:        oneToMany,
		fillRate:   0.5,
		minObjects: 1,
		maxObjects: 10,
	}, tg)
}

func TestTagParseInvalid(t *testing.T) {
	fixtureTagStr := "one-to-none,0.5,1,10"
	_, err := parseTag(fixtureTagStr)
	require.Error(t, err)
}

func TestTagParseManyInvalidFill(t *testing.T) {
	fixtureTagStr := "one-to-many,1.5,1,10"
	_, err := parseTag(fixtureTagStr)
	require.Error(t, err)
}

var expectedUserGQL = `type tUser {
	Name: String
	Age: Int
	Points: Float
	Verified: Boolean
}`

func TestExtractGQLFromTypeNoRelation(t *testing.T) {
	gql, err := ExtractGQLFromType(tUser{})
	require.NoError(t, err)
	require.Equal(t, expectedUserGQL, gql)
}

var expectedAuthorGQL = `type tAuthor {
	Name: String
	Age: Int
	Verified: Boolean
	Wrote: tBook @primary
}`

func TestExtractGQLFromTypeOneToOnePrimary(t *testing.T) {
	gql, err := ExtractGQLFromType(tAuthor{})
	require.NoError(t, err)
	require.Equal(t, expectedAuthorGQL, gql)
}

var expectedBookGQL = `type tBook {
	Name: String
	Rating: Float
	Author: tAuthor
	Publisher: tPublisher
}`

func TestExtractGQLFromTypeOneToOneSecondaryAndOneToMany(t *testing.T) {
	gql, err := ExtractGQLFromType(tBook{})
	require.NoError(t, err)
	require.Equal(t, expectedBookGQL, gql)
}

var expectedPublisherGQL = `type tPublisher {
	Name: String
	PhoneNumber: String
	FavouritePageNumbers: [Int]
	Published: [tBook]
}`

func TestExtractGQLFromTypeOneToMany(t *testing.T) {
	gql, err := ExtractGQLFromType(tPublisher{})
	require.NoError(t, err)
	require.Equal(t, expectedPublisherGQL, gql)
}
