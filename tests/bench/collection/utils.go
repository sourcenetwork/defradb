// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
				collections[k].Get(ctx, listOfDocIDs[j][k], false) //nolint:errcheck
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
					col.Get(ctx, docID, false) //nolint:errcheck
					wg.Done()
				}(ctx, collections[k], listOfDocIDs[j][k])
			}
		}

		wg.Wait()
	}
	b.StopTimer()

	return nil
}

func runCollectionBenchCreate(
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
		return runCollectionBenchCreateSync(b, ctx, collections, fixture, docCount, opCount)
	}
	return runCollectionBenchCreateAsync(b, ctx, collections, fixture, docCount, opCount)
}

func runCollectionBenchCreateMany(
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
	// CreateMany make sure numTypes == 1 since we only support that for now
	// @todo: Add support for numTypes > 1 later
	if numTypes != 1 {
		return errors.New(fmt.Sprintf("Invalid number of types for create many, have %v but max is 1", numTypes))
	}

	// run benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		docs := make([]*client.Document, opCount)
		for j := 0; j < opCount; j++ {
			d, _ := fixture.GenerateDocs()
			docs[j], _ = client.NewDocFromJSON([]byte(d[0]), collections[0].Schema())
		}

		collections[0].CreateMany(ctx, docs) //nolint:errcheck
	}
	b.StopTimer()

	return nil
}

func runCollectionBenchCreateSync(b *testing.B,
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
				doc, _ := client.NewDocFromJSON([]byte(docs[k]), collections[k].Schema())
				collections[k].Create(ctx, doc) //nolint:errcheck
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
func runCollectionBenchCreateAsync(b *testing.B,
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
					// create the documents
					for j := 0; j < numTypes; j++ {
						doc, _ := client.NewDocFromJSON([]byte(docs[j]), collections[j].Schema())
						collections[j].Create(ctx, doc) //nolint:errcheck
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
