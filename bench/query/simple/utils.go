package query

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document/key"
)

func runQueryBenchGet(b *testing.B, ctx context.Context, fixture fixtures.Generator, docCount int, query string, doSync bool) error {
	b.StopTimer()
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	dockeys, err := benchutils.BackfillBenchmarkDB(b, ctx, collections, fixture, docCount, 0, doSync)
	if err != nil {
		return err
	}

	numTypes := len(fixture.Types())
	return runQueryBenchGetSync(b, ctx, db, collections, fixture, docCount, numTypes, dockeys, query)
}

func runQueryBenchGetSync(
	b *testing.B,
	ctx context.Context,
	db *db.DB,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, numTypes int,
	dockeys [][]key.DocKey,
	query string,
) error {
	// fmt.Printf("Query:\n%s\n", query)

	// any preprocessing?
	query = formatQuery(b, query, dockeys)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res := db.ExecQuery(ctx, query)
		if len(res.Errors) > 0 {
			return fmt.Errorf("Query error: %v", res.Errors)
		}
		// l := len(res.Data.([]map[string]interface{}))
		// if l != opCount {
		// 	return fmt.Errorf("Invalid response, returned data doesn't match length, expected %v actual %v", docCount, l)
		// }
		// fmt.Println(res)
		// fmt.Println("--------------------")
	}
	return nil
}

func formatQuery(b *testing.B, query string, dockeys [][]key.DocKey) string {
	numPlaceholders := strings.Count(query, "{{dockey}}")
	if numPlaceholders == 0 {
		return query
	}
	// create a copy of dockeys since we'll be mutating it
	dockeysCopy := dockeys[:]

	// b.Logf("formatting query, replacing %v instances", numPlaceholders)
	// b.Logf("Query before: %s", query)

	if len(dockeysCopy) < numPlaceholders {
		b.Fatalf("Invalid number of query placeholders, max is %v requested is %v", len(dockeys), numPlaceholders)
	}

	for i := 0; i < numPlaceholders; i++ {
		// pick a random dockey, needs to be unique accross all
		// calls, so remove the selected one between calls
		rIndex := rand.Intn(len(dockeysCopy))
		key := dockeysCopy[rIndex][0]

		// remove selected key
		dockeysCopy = append(dockeysCopy[:rIndex], dockeysCopy[rIndex+1:]...)

		// replace
		query = strings.Replace(query, "{{dockey}}", key.String(), 1)
	}

	// b.Logf("Query After: %s", query)
	return query
}
