package fixtures

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// #2
type tBook struct {
	Name        string      `faker:"title"`
	Rating      float32     `faker:"amount"`
	Author      *tAuthor    `fixture:"one-to-one"`
	Publisher   *tPublisher `fixture:"one-to-many"`
	PublisherId string
}

// #3
type tAuthor struct {
	Name     string `faker:"name"`
	Age      int
	Verified bool
	Wrote    *tBook `fixture:"one-to-one,primary,0.8"`
	WorteId  string
}

// #1
type tPublisher struct {
	Name                 string `faker:"title"`
	PhoneNumber          string `faker:"phone_number"`
	FavouritePageNumbers []int

	// Fixture Data:
	// Rate: 50%
	// Min:1
	// Max:10
	Published []*Book `fixture:"one-to-many,0.5,1,10"`
}

func TestDependantsOneToOneSecondaryWithOneToManyPrimary(t *testing.T) {
	b := tBook{}
	val := reflect.TypeOf(b)

	deps, err := dependants(val)
	require.NoError(t, err)
	require.Equal(t, []string{"tPublisher"}, deps)
}

func TestDependantsOneToOnePrimary(t *testing.T) {
	b := tAuthor{}
	val := reflect.TypeOf(b)

	deps, err := dependants(val)
	require.NoError(t, err)
	require.Equal(t, []string{"tBook"}, deps)
}

func TestDependantsOneToManySeconday(t *testing.T) {
	b := tPublisher{}
	val := reflect.TypeOf(b)

	deps, err := dependants(val)
	require.NoError(t, err)
	require.Equal(t, []string{}, deps)
}

func TestDependantsGraph(t *testing.T) {
	types := []interface{}{
		tPublisher{},
		tAuthor{},
		tBook{},
	}

	var basicGraph Graph
	for _, v := range types {
		typ := reflect.TypeOf(v)
		deps, err := dependants(typ)
		require.NoError(t, err)

		name := typ.Name()
		basicGraph = append(basicGraph, NewNode(name, deps...))
	}

	resolvedGraph, err := resolveGraph(basicGraph)
	require.NoError(t, err)

	for _, n := range resolvedGraph {
		fmt.Println(n.name)
	}

}
