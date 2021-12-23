package collection

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
	ctx := context.Background()
	db, collections, err := setupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	dockeys, err := backfillBenchmarkDB(b, ctx, collections, fixture, docCount, opCount, doSync)
	if err != nil {
		return err
	}

	numTypes := len(fixture.Types())

	// run benchmark
	b.StartTimer()
	if doSync {
		return runCollectionBenchGetSync(b, ctx, collections, fixture, docCount, opCount, numTypes, dockeys)
	}
	return runCollectionBenchGetAsync(b, ctx, collections, fixture, docCount, opCount, numTypes, dockeys)
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

		wg.Wait()
	}

	return nil
}

func runCollectionBenchCreate(b *testing.B, fixture fixtures.Context, docCount, opCount int, doSync bool) error {
	b.StopTimer()
	ctx := context.Background()
	db, collections, err := setupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = backfillBenchmarkDB(b, ctx, collections, fixture, docCount, opCount, doSync)
	if err != nil {
		return err
	}

	numTypes := len(fixture.Types())

	// run benchmark
	b.StartTimer()
	if doSync {
		return runCollectionBenchCreateSync(b, ctx, collections, fixture, docCount, opCount, numTypes)
	}
	return nil
}

func setupDBAndCollections(b *testing.B, ctx context.Context, fixture fixtures.Context) (*defradb.DB, []client.Collection, error) {
	db, err := testutils.NewTestDB()
	if err != nil {
		return nil, nil, err
	}

	// create collections
	collections, err := setupCollections(b, ctx, db, fixture)
	if err != nil {
		return nil, nil, err
	}

	return db, collections, nil

}

func backfillBenchmarkDB(b *testing.B, ctx context.Context, cols []client.Collection, fixture fixtures.Context, docCount, opCount int, doSync bool) ([][]key.DocKey, error) {
	numTypes := len(fixture.Types())

	// load fixtures
	var wg sync.WaitGroup
	wg.Add(docCount)
	errCh := make(chan error)
	waitCh := make(chan struct{})
	dockeys := make([][]key.DocKey, docCount)

	go func() {
		for i := 0; i < docCount; i++ {
			go func(index int) {
				docs, err := fixture.GenerateDocs()
				if err != nil {
					errCh <- fmt.Errorf("Failed to generate document payload from fixtures: %w", err)
					return
				}

				fmt.Println(docs)

				// create the documents
				keys := make([]key.DocKey, numTypes)
				for j := 0; j < numTypes; j++ {

					doc, err := document.NewFromJSON([]byte(docs[j]))
					if err != nil {
						errCh <- fmt.Errorf("Failed to create document from fixture: %w", err)
						return
					}

					if err := cols[j].Create(ctx, doc); err != nil {
						fmt.Println("Failed on", doc)
						errCh <- fmt.Errorf("Failed to create document on collection: %w", err)
						return
					}
					fmt.Println("created:", doc)
					keys[j] = doc.Key()
				}
				dockeys[index] = keys

				wg.Done()
			}(i)
		}

		// wait for our group and signal by closing waitCh
		wg.Wait()
		close(waitCh)
	}()

	// finish or err
	select {
	case <-waitCh:
		return dockeys, nil
	case err := <-errCh:
		return nil, err
	}

}

func runCollectionBenchCreateSync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Context,
	docCount, opCount, numTypes int,
) error {

	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount/numTypes; j++ {
			for k := 0; k < numTypes; k++ {
				docs, _ := fixture.GenerateDocs()

				// create the documents
				for h := 0; h < numTypes; h++ {
					doc, _ := document.NewFromJSON([]byte(docs[h]))
					collections[k].Create(ctx, doc)
				}
			}
		}
	}

	return nil
}
