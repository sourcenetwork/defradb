// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package storage

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/client"
)

func runStorageBenchGet(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint

	// backfill
	keys, err := backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			positionInInterval := getSampledIndex(len(keys), opCount, j)
			key := ds.NewKey(keys[positionInInterval])
			_, err := db.Get(ctx, key)
			if err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchTxnGet(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := benchutils.NewTestDB(ctx, b)

	if err != nil {
		return err
	}
	defer db.Root().Close() //nolint

	keys, err := backfillBenchmarkTxn(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			positionInInterval := getSampledIndex(len(keys), opCount, j)
			key := ds.NewKey(keys[positionInInterval])
			_, err := txn.Rootstore().Get(ctx, key)
			if err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchTxnIterator(b *testing.B, ctx context.Context, valueSize, objCount, opCount, pointCount int, doSync bool) error {
	db, err := benchutils.NewTestDB(ctx, b)

	if err != nil {
		return err
	}
	defer db.Root().Close() //nolint

	keys, err := backfillBenchmarkStorageDB(ctx, db.Root(), objCount, valueSize)
	if err != nil {
		return err
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			iterator, err := txn.Rootstore().GetIterator(query.Query{})
			if err != nil {
				return err
			}
			totalCount := 0
			b.StartTimer()

			for k := 0; k < pointCount; k++ {
				positionInInterval := getSampledIndex(len(keys), pointCount, k)
				startKey := ds.NewKey(keys[positionInInterval])

				result, err := iterator.IteratePrefix(ctx, startKey, startKey)
				if err != nil {
					return err
				}
				resCount := 0
				for {
					_, hasNextItem := result.NextSync()
					if !hasNextItem {
						break
					}
					totalCount++
					resCount++
				}
				// err = result.Close()
				// if err != nil {
				// 	return err
				// }
			}
			b.StopTimer()
			err = iterator.Close()
			if err != nil {
				return err
			}

			// fmt.Println("COUNT:", totalCount, pointCount)
		}
	}
	b.StopTimer()
	txn.Discard(ctx)
	return nil
}

func runStorageBenchTxnIteratorRaw(b *testing.B, ctx context.Context, valueSize, objCount int, keyonly bool) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close()

	// backfill
	_, err = backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	dbtxn, ok := db.(ds.TxnDatastore)
	if !ok {
		return errors.New("failed to get Txn Datastore from Test Storage backend")
	}

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		txn, err := dbtxn.NewTransaction(ctx, true)
		if err != nil {
			return err
		}
		b.StartTimer()

		// iterate over all keys
		res, err := txn.Query(ctx, dsq.Query{
			// Prefix: "/data",
			KeysOnly: keyonly,
		})
		if err != nil {
			return err
		}

		resCount := 0
		for {
			_, hasNext := res.NextSync()
			if !hasNext {
				break
			}
			// panic("hi andy")
			// fmt.Println("===", resCount, e.Key, "||||", string(e.Value), "...")
			resCount++
		}

		// if resCount != objCount+2 {
		// 	return fmt.Errorf("incorrect query iterator doc count, expected %v got %v", objCount, resCount)
		// }

		b.StopTimer()
		txn.Discard(ctx)
	}

	return nil
}

func runStorageBenchTxnIteratorRawSkip(b *testing.B, ctx context.Context, valueSize, objCount int, skip float32) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close()

	// backfill
	_, err = backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	dbtxn, ok := db.(ds.TxnDatastore)
	if !ok {
		return errors.New("failed to get Txn Datastore from Test Storage backend")
	}

	threshold := int32(100 - (100 * skip))

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		txn, err := dbtxn.NewTransaction(ctx, true)
		if err != nil {
			return err
		}
		b.StartTimer()

		// iterate over all keys
		res, err := txn.Query(ctx, dsq.Query{
			// Prefix: "/data",
			KeysOnly: true,
		})
		if err != nil {
			return err
		}

		resCount := 0
		for {
			e, hasNext := res.NextSync()
			if !hasNext {
				break
			}

			// dont skip?
			if rand.Int31n(100) < threshold {
				_, err = txn.Get(ctx, ds.NewKey(e.Key))
				if err != nil {
					return err
				}
			}
			// fmt.Println("===", resCount, e.Key, "||||", string(e.Value), "...")

			resCount++
		}

		// if resCount != objCount {
		// 	return fmt.Errorf("incorrect query iterator doc count, expected %v got %v", objCount, resCount)
		// }

		b.StopTimer()
		txn.Discard(ctx)
	}

	return nil
}

func runStorageBenchPut(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint

	// backfill
	_, err = backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			keyBuf := make([]byte, 32)
			value := make([]byte, valueSize)
			if _, err := rand.Read(value); err != nil {
				return err
			}
			if _, err := rand.Read(keyBuf); err != nil {
				return err
			}
			key := ds.NewKey(string(keyBuf))

			if err := db.Put(ctx, key, value); err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchPutMany(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint

	// backfill
	_, err = backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	//shuffle keys
	// rand.Shuffle(len(keys), func(i, j int) {
	// 	keys[i], keys[j] = keys[j], keys[i]
	// })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch, err := db.Batch(ctx)
		if err != nil {
			return err
		}
		for j := 0; j < opCount; j++ {
			keyBuf := make([]byte, 32)
			value := make([]byte, valueSize)
			if _, err := rand.Read(value); err != nil {
				return err
			}
			if _, err := rand.Read(keyBuf); err != nil {
				return err
			}
			key := ds.NewKey(string(keyBuf))

			if err := batch.Put(ctx, key, value); err != nil {
				return err
			}
		}
		if err := batch.Commit(ctx); err != nil {
			return err
		}
	}
	b.StopTimer()

	return nil
}

func backfillBenchmarkStorageDB(ctx context.Context, db ds.Batching, objCount int, valueSize int) ([]string, error) {
	batch, err := db.Batch(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]string, objCount)
	for i := 0; i < objCount; i++ {
		// keyBuf := make([]byte, 32)
		// value := make([]byte, valueSize)
		// if _, err := rand.Read(value); err != nil {
		// 	return nil, err
		// }
		// if _, err := rand.Read(keyBuf); err != nil {
		// 	return nil, err
		// }
		keyBuf := randSeq(32)
		value := []byte(randSeq(valueSize))
		key := ds.NewKey("/data" + string(keyBuf))
		keys[i] = key.String()

		if err := batch.Put(ctx, key, value); err != nil {
			return nil, err
		}
	}

	sort.Strings(keys)
	return keys, batch.Commit(ctx)
}

func backfillBenchmarkTxn(ctx context.Context, db client.DB, objCount int, valueSize int) ([]string, error) {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	keys := make([]string, objCount)
	for i := 0; i < objCount; i++ {
		// keyBuf := make([]byte, 32)
		// value := make([]byte, valueSize)
		// if _, err := rand.Read(value); err != nil {
		// 	return nil, err
		// }
		// if _, err := rand.Read(keyBuf); err != nil {
		// 	return nil, err
		// }
		keyBuf := randSeq(32)
		value := []byte(randSeq(valueSize))
		key := ds.NewKey("/data" + string(keyBuf))
		keys[i] = key.String()

		if err := txn.Rootstore().Put(ctx, key, value); err != nil {
			return nil, err
		}
	}

	sort.Strings(keys)
	return keys, txn.Commit(ctx)
}

func getSampledIndex(populationSize int, sampleSize int, i int) int {
	if sampleSize >= populationSize {
		if i == 0 {
			return 0
		}
		return (populationSize - 1) / i
	}

	pointsPerInterval := populationSize / sampleSize
	return (i * pointsPerInterval) + rand.Intn(pointsPerInterval)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
