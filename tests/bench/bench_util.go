// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package bench

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"

	"github.com/dgraph-io/badger/v4"
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
	testutils "github.com/sourcenetwork/defradb/tests/integration"
)

const (
	writeBatchGroup = 100
	storageEnvName  = "DEFRA_BENCH_STORAGE"
)

var (
	storage string = "memory"
	log            = logging.MustNewLogger("tests.bench")
)

func init() {
	logging.SetConfig(logging.Config{Level: logging.NewLogLevelOption(logging.Error)})

	// assign if not empty
	if s := os.Getenv(storageEnvName); s != "" {
		storage = s
	}
}

func SetupCollections(
	b *testing.B,
	ctx context.Context,
	db client.DB,
	fixture fixtures.Generator,
) ([]client.Collection, error) {
	numTypes := len(fixture.Types())
	collections := make([]client.Collection, numTypes)
	schema, err := ConstructSchema(fixture)
	if err != nil {
		return nil, err
	}

	// b.Logf("Loading schema: \n%s", schema)

	if _, err := db.AddSchema(ctx, schema); err != nil {
		return nil, errors.Wrap("couldn't load schema", err)
	}

	// loop to get collections
	for i := 0; i < numTypes; i++ {
		col, err := db.GetCollectionByName(ctx, fixture.TypeName(i))
		if err != nil {
			return nil, errors.Wrap(fmt.Sprintf("Couldn't get the collection %v", fixture.TypeName(i)), err)
		}
		// b.Logf("Collection Name: %s", col.Name())
		collections[i] = col
	}

	return collections, nil
}

func ConstructSchema(fixture fixtures.Generator) (string, error) {
	numTypes := len(fixture.Types())
	var schema string

	// loop to get the schemas
	for i := 0; i < numTypes; i++ {
		gql, err := fixtures.ExtractGQLFromType(fixture.Types()[i])
		if err != nil {
			return "", errors.Wrap("failed generating GQL", err)
		}

		schema += gql
		schema += "\n\n"
	}

	return schema, nil
}

func SetupDBAndCollections(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
) (client.DB, []client.Collection, error) {
	db, err := NewTestDB(ctx, b)
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

// Loads the given test database using the provided fixture context.
// It loads docCount number of documents asynchronously in batches of *up to*
// writeBatchGroup.
func BackfillBenchmarkDB(
	b *testing.B,
	ctx context.Context,
	cols []client.Collection,
	fixture fixtures.Generator,
	docCount, opCount int,
	doSync bool,
) ([][]client.DocKey, error) {
	numTypes := len(fixture.Types())

	// load fixtures
	var wg sync.WaitGroup
	wg.Add(docCount)
	errCh := make(chan error)
	waitCh := make(chan struct{})
	dockeys := make([][]client.DocKey, docCount)

	go func() {
		// Cut up the job from into writeBatchGroup size grouped jobs.
		// Note weird math because the last batch will likely be smaller then
		// writeBatchGroup ~cus math~.
		for bid := 0; float64(bid) < math.Ceil(float64(docCount)/writeBatchGroup); bid++ {
			currentBatchSize := int(
				math.Min(float64((docCount - (bid * writeBatchGroup))), writeBatchGroup),
			)
			var batchWg sync.WaitGroup
			batchWg.Add(currentBatchSize)

			// spin up a goroutine for each doc in the current batch.
			// wait for the entire batch to finish before moving on to
			// the next batch
			for i := 0; i < currentBatchSize; i++ {
				go func(index int) {
					docs, err := fixture.GenerateDocs()
					if err != nil {
						errCh <- errors.Wrap("failed to generate document payload from fixtures", err)
						return
					}

					// create the documents
					keys := make([]client.DocKey, numTypes)
					for j := 0; j < numTypes; j++ {
						doc, err := client.NewDocFromJSON([]byte(docs[j]))
						if err != nil {
							errCh <- errors.Wrap("failed to create document from fixture", err)
							return
						}

						// loop forever until committed.
						// This was necessary when debugging and was left
						// in place. The error check could prob use a wrap system
						// but its fine :).
						for {
							if err := cols[j].Create(ctx, doc); err != nil &&
								err.Error() == badger.ErrConflict.Error() {
								log.Info(
									ctx,
									"Failed to commit TX for doc %s, retrying...\n",
									logging.NewKV("DocKey", doc.Key()),
								)
								continue
							} else if err != nil {
								errCh <- errors.Wrap("failed to create document", err)
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

func NewTestDB(ctx context.Context, t testing.TB) (client.DB, error) {
	dbi, err := newBenchStoreInfo(ctx, t)
	return dbi, err
}

func NewTestStorage(ctx context.Context, t testing.TB) (ds.Batching, error) {
	dbi, err := newBenchStoreInfo(ctx, t)
	return dbi.Root(), err
}

func newBenchStoreInfo(ctx context.Context, t testing.TB) (client.DB, error) {
	var db client.DB
	var err error

	switch storage {
	case "memory":
		db, err = testutils.NewBadgerMemoryDB(ctx)
	case "badger":
		db, _, err = testutils.NewBadgerFileDB(ctx, t)
	default:
		return nil, errors.New(fmt.Sprintf("invalid storage engine backend: %s", storage))
	}

	if err != nil {
		return nil, errors.Wrap("failed to create storage backend", err)
	}
	return db, err
}
