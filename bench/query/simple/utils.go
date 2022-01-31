package query

import (
	"context"
	"fmt"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document/key"
)

func runQueryBenchGet(b *testing.B, fixture fixtures.Context, docCount int, query string, doSync bool) error {
	b.StopTimer()
	ctx := context.Background()
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
	fixture fixtures.Context,
	docCount, numTypes int,
	dockeys [][]key.DocKey,
	query string,
) error {
	// fmt.Printf("Query:\n%s\n", query)
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
