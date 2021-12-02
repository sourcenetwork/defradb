package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	defradb "github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

func Benchmark_Collection_UserSimpleOne_Read_1_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1, 1)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_10_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 10, 10)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_100_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 100, 100)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_1000(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, 1000)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_1(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, 1)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_10(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, 10)
	if err != nil {
		b.Fatal(err)
	}
}

func Benchmark_Collection_UserSimpleOne_Read_1000_100(b *testing.B) {
	ctx := context.Background()
	err := runCollectionBenchGet(b, fixtures.WithSchema(ctx, "user_simple"), 1000, 100)
	if err != nil {
		b.Fatal(err)
	}
}

func setupCollections(b *testing.B, db *defradb.DB, ctx fixtures.Context) ([]client.Collection, error) {
	// create collection
	numTypes := len(ctx.Types())
	collections := make([]client.Collection, numTypes)
	var schema string

	// loop to get the schemas
	for i := 0; i < numTypes; i++ {
		gql, err := fixtures.ExtractGQLFromType(ctx.Types()[i])
		if err != nil {
			return nil, fmt.Errorf("failed generating GQL: %w", err)
		}

		schema += gql
		schema += "\n\n"
	}

	// b.Logf("Loading schema: \n%s", schema)

	if err := db.AddSchema(schema); err != nil {
		return nil, fmt.Errorf("Couldn't load schema: %w", err)
	}

	// loop to get collections
	for i := 0; i < numTypes; i++ {
		col, err := db.GetCollection(ctx.TypeName(i))
		if err != nil {
			return nil, fmt.Errorf("Couldn't get the collection %v: %w", ctx.TypeName(i), err)
		}
		collections[i] = col
	}

	return collections, nil
}

func runCollectionBenchGet(b *testing.B, ctx fixtures.Context, docCount, opCount int) error {
	b.StopTimer()

	db, err := defradb.NewDB(dbopts)
	if err != nil {
		return err
	}
	defer cleanupDB(db)

	// create collections
	numTypes := len(ctx.Types())
	collections, err := setupCollections(b, db, ctx)
	if err != nil {
		return err
	}

	// load fixtures
	dockeys := make([][]key.DocKey, docCount)
	for i := 0; i < docCount; i++ {
		docs, err := ctx.GenerateFixtureDocs()
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

			if err := collections[j].Create(doc); err != nil {
				return fmt.Errorf("Failed to create document on collection: %w", err)
			}
			keys[j] = doc.Key()
		}
		dockeys[i] = keys
	}

	// run benchmark
	b.StartTimer()

	for i := 0; i < b.N; i++ { // outer benchmark loop
		for j := 0; j < opCount/numTypes; j++ { // number of Get operations we want to execute
			for k := 0; k < numTypes; k++ { // apply op to all the related types
				collections[k].Get(dockeys[j][k])
			}
		}
	}

	return nil
}
