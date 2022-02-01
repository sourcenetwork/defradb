package collection

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

const (
	writeBatchGroup = 100
)

func runCollectionBenchGet(b *testing.B, ctx context.Context, fixture fixtures.Generator, docCount, opCount int, doSync bool) error {
	b.StopTimer()
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	dockeys, err := benchutils.BackfillBenchmarkDB(b, ctx, collections, fixture, docCount, opCount, doSync)
	if err != nil {
		return err
	}

	// fmt.Println("Finished backfill...")
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
	fixture fixtures.Generator,
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

// pretty basic async loop, one goroutine for
// each operation we need to do
func runCollectionBenchGetAsync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
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

func runCollectionBenchCreate(b *testing.B, ctx context.Context, fixture fixtures.Generator, docCount, opCount int, doSync bool) error {
	b.StopTimer()
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = benchutils.BackfillBenchmarkDB(b, ctx, collections, fixture, docCount, opCount, doSync)
	if err != nil {
		return err
	}

	numTypes := len(fixture.Types())

	// run benchmark
	b.StartTimer()
	if doSync {
		return runCollectionBenchCreateSync(b, ctx, collections, fixture, docCount, opCount, numTypes)
	}
	return runCollectionBenchCreateAsync2(b, ctx, collections, fixture, docCount, opCount, numTypes)
}

func runCollectionBenchCreateMany(b *testing.B, ctx context.Context, fixture fixtures.Generator, docCount, opCount int, doSync bool) error {
	b.StopTimer()
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = benchutils.BackfillBenchmarkDB(b, ctx, collections, fixture, docCount, opCount, doSync)
	if err != nil {
		return err
	}

	numTypes := len(fixture.Types())
	// CreateMany make sure numTypes == 1 since we only support that for now
	// @todo: Add support for numTypes > 1 later
	if numTypes != 1 {
		return fmt.Errorf("Invalid number of types for create many, have %v but max is 1", numTypes)
	}

	// run benchmark

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		docs := make([]*document.Document, opCount)
		for j := 0; j < opCount; j++ {
			d, _ := fixture.GenerateDocs()
			docs[j], _ = document.NewFromJSON([]byte(d[0]))
		}

		collections[0].CreateMany(ctx, docs)
	}

	return nil
}

func runCollectionBenchCreateSync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount, numTypes int,
) error {

	runs := opCount / numTypes
	for i := 0; i < b.N; i++ {
		for j := 0; j < runs; j++ {
			docs, _ := fixture.GenerateDocs()
			for k := 0; k < numTypes; k++ {
				doc, _ := document.NewFromJSON([]byte(docs[k]))
				collections[k].Create(ctx, doc)
			}
		}
	}

	return nil
}

// workers
// IGNORE THIS, unused, sorry andy :).
func runCollectionBenchCreateAsync1(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount, numTypes int,
) error {
	// fmt.Println("----------------------------------------------------------")
	// init the workers
	b.StopTimer()
	closeCh := make(chan struct{})
	workerCh := make(chan struct{}, writeBatchGroup)
	var wg sync.WaitGroup
	for i := 0; i < writeBatchGroup; i++ {
		go func() {
			for {
				select {
				case <-workerCh:
					docs, _ := fixture.GenerateDocs()
					for k := 0; k < numTypes; k++ {
						doc, _ := document.NewFromJSON([]byte(docs[k]))
						collections[k].Create(ctx, doc)
					}
					wg.Done()
				case <-closeCh:
					return
				}
			}
		}()
	}

	runs := opCount / numTypes
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(runs)
		for j := 0; j < runs; j++ {
			workerCh <- struct{}{} // send job notification
		}
		wg.Wait()
	}

	close(closeCh)
	return nil
}

// batching
// uses an async method similar to the BackFill implementaion
// cuts the total task up into batchs up to writeBatchGroup size
// and wait for it all to finish.
func runCollectionBenchCreateAsync2(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount, numTypes int,
) error {

	for bi := 0; bi < b.N; bi++ {
		var wg sync.WaitGroup
		wg.Add(opCount)

		for bid := 0; float64(bid) < math.Ceil(float64(opCount)/writeBatchGroup); bid++ {
			currentBatchSize := int(math.Min(float64((opCount - (bid * writeBatchGroup))), writeBatchGroup))
			var batchWg sync.WaitGroup
			batchWg.Add(currentBatchSize)

			for i := 0; i < currentBatchSize; i++ {
				go func(index int) {
					docs, _ := fixture.GenerateDocs()
					// create the documents
					for j := 0; j < numTypes; j++ {
						doc, _ := document.NewFromJSON([]byte(docs[j]))
						collections[j].Create(ctx, doc)
					}

					wg.Done()
					batchWg.Done()
				}((bid * writeBatchGroup) + i)
			}

			batchWg.Wait()
			// fmt.Printf(".")
		}

		// finish or err
		wg.Wait()
	}

	return nil
}
