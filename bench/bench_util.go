package bench

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/dgraph-io/badger/v3"

	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	defradb "github.com/sourcenetwork/defradb/db"
	testutils "github.com/sourcenetwork/defradb/db/tests"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

const (
	writeBatchGroup = 100
)

func SetupCollections(b *testing.B, ctx context.Context, db *defradb.DB, fixture fixtures.Context) ([]client.Collection, error) {
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
		// b.Logf("Collection Name: %s", col.Name())
		collections[i] = col
	}

	return collections, nil
}

func SetupDBAndCollections(b *testing.B, ctx context.Context, fixture fixtures.Context) (*defradb.DB, []client.Collection, error) {
	db, err := testutils.NewTestDB()
	if err != nil {
		return nil, nil, err
	}

	// create collections
	collections, err := SetupCollections(b, ctx, db, fixture)
	if err != nil {
		return nil, nil, err
	}

	return db, collections, nil

}

func BackfillBenchmarkDB(b *testing.B, ctx context.Context, cols []client.Collection, fixture fixtures.Context, docCount, opCount int, doSync bool) ([][]key.DocKey, error) {
	numTypes := len(fixture.Types())

	// load fixtures
	var wg sync.WaitGroup
	wg.Add(docCount)
	errCh := make(chan error)
	waitCh := make(chan struct{})
	dockeys := make([][]key.DocKey, docCount)

	go func() {
		for bid := 0; float64(bid) < math.Ceil(float64(docCount)/writeBatchGroup); bid++ {
			currentBatchSize := int(math.Min(float64((docCount - (bid * writeBatchGroup))), writeBatchGroup))
			var batchWg sync.WaitGroup
			batchWg.Add(currentBatchSize)

			for i := 0; i < currentBatchSize; i++ {
				go func(index int) {
					docs, err := fixture.GenerateDocs()
					if err != nil {
						errCh <- fmt.Errorf("Failed to generate document payload from fixtures: %w", err)
						return
					}

					// fmt.Println(docs)

					// create the documents
					keys := make([]key.DocKey, numTypes)
					for j := 0; j < numTypes; j++ {

						doc, err := document.NewFromJSON([]byte(docs[j]))
						if err != nil {
							errCh <- fmt.Errorf("Failed to create document from fixture: %w", err)
							return
						}

						for { // loop untill commited
							if err := cols[j].Create(ctx, doc); err != nil && err.Error() == badger.ErrConflict.Error() {
								fmt.Printf("failed to commit TX for doc %s, retrying...\n", doc.Key())
								continue
							} else if err != nil {
								errCh <- fmt.Errorf("Failed to create document: %w", err)
							}
							keys[j] = doc.Key()
							break
						}
					}
					dockeys[index] = keys

					wg.Done()
					batchWg.Done()
				}((bid * writeBatchGroup) + i)
			}

			batchWg.Wait()
			// fmt.Printf(".")
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
