// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package collection

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	benchutils "github.com/sourcenetwork/defradb/tests/bench"
	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

const (
	writeBatchGroup = 100
)

func runCollectionBenchGet(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	docCount, opCount int,
	doSync bool,
) error {
	db, collections, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	listOfDocIDs, err := benchutils.BackfillBenchmarkDB(
		b,
		ctx,
		collections,
		fixture,
		docCount,
		opCount,
		doSync,
	)
	if err != nil {
		return err
	}

	// run benchmark
	if doSync {
		return runCollectionBenchGetSync(b, ctx, collections, fixture, docCount, opCount, listOfDocIDs)
	}
	return runCollectionBenchGetAsync(b, ctx, collections, fixture, docCount, opCount, listOfDocIDs)
}

func runCollectionBenchGetSync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount int,
	listOfDocIDs [][]client.DocID,
) error {
	numTypes := len(fixture.Types())
	b.ResetTimer()
	for i := 0; i < b.N; i++ { // outer benchmark loop
		for j := 0; j < opCount/numTypes; j++ { // number of Get operations we want to execute
			for k := 0; k < numTypes; k++ { // apply op to all the related types
				collections[k].GetDocument( //nolint:errcheck
					ctx,
					listOfDocIDs[j][k],
				)
			}
		}
	}
	b.StopTimer()

	return nil
}

// pretty basic async loop, one goroutine for
// each operation we need to do
func runCollectionBenchGetAsync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount int,
	listOfDocIDs [][]client.DocID,
) error {
	var wg sync.WaitGroup
	numTypes := len(fixture.Types())
	b.ResetTimer()
	for i := 0; i < b.N; i++ { // outer benchmark loop
		for j := 0; j < opCount/numTypes; j++ { // number of Get operations we want to execute
			for k := 0; k < numTypes; k++ { // apply op to all the related types
				wg.Add(1)
				go func(ctx context.Context, col client.Collection, docID client.DocID) {
					col.GetDocument( //nolint:errcheck
						ctx,
						docID,
					)
					wg.Done()
				}(ctx, collections[k], listOfDocIDs[j][k])
			}
		}

		wg.Wait()
	}
	b.StopTimer()

	return nil
}

func runCollectionBenchAdd(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	docCount, opCount int,
	doSync bool,
) error {
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

	// run benchmark
	b.StartTimer()
	if doSync {
		return runCollectionBenchAddSync(b, ctx, collections, fixture, docCount, opCount)
	}
	return runCollectionBenchAddAsync(b, ctx, collections, fixture, docCount, opCount)
}

func runCollectionBenchAddMany(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	docCount, opCount int,
	doSync bool,
) error {
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
	// AddMany make sure numTypes == 1 since we only support that for now
	// @todo: Add support for numTypes > 1 later
	if numTypes != 1 {
		return errors.New(fmt.Sprintf("Invalid number of types for add many, have %v but max is 1", numTypes))
	}

	// run benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docs := make([]*client.Document, opCount)
		for j := 0; j < opCount; j++ {
			d, _ := fixture.GenerateDocs()
			docs[j], _ = client.NewDocFromJSON(ctx, []byte(d[0]), collections[0].Version())
		}

		collections[0].AddManyDocuments(ctx, docs) //nolint:errcheck
	}
	b.StopTimer()

	return nil
}

func runCollectionBenchAddSync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount int,
) error {
	numTypes := len(fixture.Types())
	b.ResetTimer()
	runs := opCount / numTypes
	for i := 0; i < b.N; i++ {
		for j := 0; j < runs; j++ {
			docs, _ := fixture.GenerateDocs()
			for k := 0; k < numTypes; k++ {
				doc, _ := client.NewDocFromJSON(ctx, []byte(docs[k]), collections[k].Version())
				collections[k].AddDocument(ctx, doc) //nolint:errcheck
			}
		}
	}
	b.StopTimer()

	return nil
}

// batching
// uses an async method similar to the BackFill implementaion
// cuts the total task up into batchs up to writeBatchGroup size
// and wait for it all to finish.
func runCollectionBenchAddAsync(b *testing.B,
	ctx context.Context,
	collections []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount int,
) error {
	numTypes := len(fixture.Types())
	b.StartTimer()

	for bi := 0; bi < b.N; bi++ {
		var wg sync.WaitGroup
		wg.Add(opCount)

		for bid := 0; float64(bid) < math.Ceil(float64(opCount)/writeBatchGroup); bid++ {
			currentBatchSize := int(
				math.Min(float64((opCount - (bid * writeBatchGroup))), writeBatchGroup),
			)
			var batchWg sync.WaitGroup
			batchWg.Add(currentBatchSize)

			for i := 0; i < currentBatchSize; i++ {
				go func(index int) {
					docs, _ := fixture.GenerateDocs()
					// add the documents
					for j := 0; j < numTypes; j++ {
						doc, _ := client.NewDocFromJSON(ctx, []byte(docs[j]), collections[j].Version())
						collections[j].AddDocument(ctx, doc) //nolint:errcheck
					}

					wg.Done()
					batchWg.Done()
				}((bid * writeBatchGroup) + i)
			}

			batchWg.Wait()
		}

		// finish or err
		wg.Wait()
	}

	return nil
}
