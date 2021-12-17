package bench

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	defradb "github.com/sourcenetwork/defradb/db"
	testutils "github.com/sourcenetwork/defradb/db/tests"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

func Benchmark_Collection_UserSimpleOne_Read_Sync_1_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_10_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 10, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_100_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 100, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_1000(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1000, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 10, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Sync_1000_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 100, true)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_1_1(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_10_10(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 10, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_100_100(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 100, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_Async_1000_1000(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1000, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_1_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 1, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_10_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 10, false)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_100_Async(b *testing.B) {
	fixture := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(fixture, "user_simple"), 1000, 100, false)
	if err != nil {
		b.Fatal(err)
	}
}

func setupCollections(b *testing.B, ctx context.Context, db *defradb.DB, fixture fixtures.Context) ([]client.Collection, error) {
	// create collection
	numTypes := len(fixture.Types())
	collections := make([]client.Collection, numTypes)
	var schema string

	// loop to get the schemas
	for i := 0; i < numTypes; i++ {
		gql, err := fixtures.ExtractGQLFromType(fixture.Types()[i])
		if err != nil {
			return nil, fmt.Errorf("failed generating GQL: %w", err)
		}

		schema += gql
		schema += "\n\n"
	}

	// b.Logf("Loading schema: \n%s", schema)

	if err := db.AddSchema(ctx, schema); err != nil {
		return nil, fmt.Errorf("Couldn't load schema: %w", err)
	}

	// loop to get collections
	for i := 0; i < numTypes; i++ {
		col, err := db.GetCollection(ctx, fixture.TypeName(i))
		if err != nil {
			return nil, fmt.Errorf("Couldn't get the collection %v: %w", fixture.TypeName(i), err)
		}
		collections[i] = col
	}

	return collections, nil
}

func runCollectionBenchGet(b *testing.B, fixture fixtures.Context, docCount, opCount int, doSync bool) error {
	b.StopTimer()

	db, err := testutils.NewTestDB()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	// create collections
	numTypes := len(fixture.Types())
	collections, err := setupCollections(b, ctx, db, fixture)
	if err != nil {
		return err
	}

	// load fixtures
	dockeys := make([][]key.DocKey, docCount)
	for i := 0; i < docCount; i++ {
		docs, err := fixture.GenerateFixtureDocs()
		if err != nil {
			return fmt.Errorf("Failed to generate document payload from fixtures: %w", err)
		}
		// @todo: Handle linking doc types for relations

		// create the documents
		keys := make([]key.DocKey, numTypes)
		for j := 0; j < numTypes; j++ {
			// b.Logf("Generated Doc:\n%s", docs[j])
			doc, err := document.NewFromJSON([]byte(docs[j]))
			if err != nil {
				return fmt.Errorf("Failed to create document from fixture: %w", err)
			}

			if err := collections[j].Create(ctx, doc); err != nil {
				return fmt.Errorf("Failed to create document on collection: %w", err)
			}
			keys[j] = doc.Key()
		}
		dockeys[i] = keys
	}

	// run benchmark
	b.StartTimer()
	if doSync {
		return runCollectionBenchGetSync(b, ctx, collections, fixture, docCount, opCount, numTypes, dockeys)
	}
	return runCollectionBenchGetAsync(b, ctx, collections, fixture, docCount, opCount, numTypes, dockeys)

	return nil
}

func runCollectionBenchGetSync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Context,
	docCount, opCount, numTypes int,
	dockeys [][]key.DocKey,
) error {

	for i := 0; i < b.N; i++ { // outer benchmark loop
		for j := 0; j < opCount/numTypes; j++ { // number of Get operations we want to execute
			for k := 0; k < numTypes; k++ { // apply op to all the related types
				collections[k].Get(ctx, dockeys[j][k])
			}
		}
	}

	return nil
}

func runCollectionBenchGetAsync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Context,
	docCount, opCount, numTypes int,
	dockeys [][]key.DocKey,
) error {

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ { // outer benchmark loop
		for j := 0; j < opCount/numTypes; j++ { // number of Get operations we want to execute
			for k := 0; k < numTypes; k++ { // apply op to all the related types
				wg.Add(1)
				go func(ctx context.Context, col client.Collection, dockey key.DocKey) {
					col.Get(ctx, dockey)
					wg.Done()
				}(ctx, collections[k], dockeys[j][k])
			}
		}
	}

	wg.Wait()
	return nil
}
