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
	"crypto/rand"
	"errors"
	mathRand "math/rand"
	"sort"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	benchutils "github.com/sourcenetwork/defradb/tests/bench"
)

func runStorageBenchGet(
	b *testing.B,
	ctx context.Context,
	valueSize, objCount, opCount int,
	doSync bool,
) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	// backfill
	keys, err := backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			positionInInterval := getSampledIndex(len(keys), opCount, j)
			key := keys[positionInInterval]
			_, err := db.Get(ctx, []byte(key))
			if err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchTxnGet(
	b *testing.B,
	ctx context.Context,
	valueSize, objCount, opCount int,
	doSync bool,
) error {
	db, err := benchutils.NewTestDB(ctx, b)

	if err != nil {
		return err
	}
	defer db.Close()

	keys, err := backfillBenchmarkTxn(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	clientTxn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer clientTxn.Discard(ctx)

	txn := txnctx.MustGetFromClient(clientTxn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			positionInInterval := getSampledIndex(len(keys), opCount, j)
			key := []byte(keys[positionInInterval])
			_, err := txn.Rootstore().Get(ctx, key)
			if err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchTxnIterator(
	b *testing.B,
	ctx context.Context,
	valueSize, objCount, opCount, pointCount int,
	doSync bool,
) error {
	db, err := benchutils.NewTestDB(ctx, b)

	if err != nil {
		return err
	}
	defer db.Close()

	keys, err := backfillBenchmarkTxn(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	clientTxn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer clientTxn.Discard(ctx)

	txn := txnctx.MustGetFromClient(clientTxn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			for k := 0; k < pointCount; k++ {
				positionInInterval := getSampledIndex(len(keys), pointCount, k)
				startKey := ds.NewKey(keys[positionInInterval])

				iter, err := txn.Rootstore().Iterator(ctx, corekv.IterOptions{
					Prefix: startKey.Bytes(),
				})
				if err != nil {
					return err
				}
				for {
					hasNextItem, err := iter.Next()
					if err != nil {
						return errors.Join(err, iter.Close())
					}
					if !hasNextItem {
						break
					}
				}
				err = iter.Close()
				if err != nil {
					return err
				}
			}
		}
	}
	b.StopTimer()
	return nil
}

func runStorageBenchPut(
	b *testing.B,
	ctx context.Context,
	valueSize, objCount, opCount int,
	doSync bool,
) error {
	db, err := benchutils.NewTestStorage(ctx, b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	// backfill
	_, err = backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			key := make([]byte, 32)
			value := make([]byte, valueSize)
			if _, err := rand.Read(value); err != nil {
				return err
			}
			if _, err := rand.Read(key); err != nil {
				return err
			}

			if err := db.Set(ctx, key, value); err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func backfillBenchmarkStorageDB(
	ctx context.Context,
	db corekv.TxnStore,
	objCount int,
	valueSize int,
) ([]string, error) {
	keys := make([]string, objCount)
	for i := 0; i < objCount; i++ {
		key := make([]byte, 32)
		value := make([]byte, valueSize)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		keys[i] = string(key)

		if err := db.Set(ctx, key, value); err != nil {
			return nil, err
		}
	}

	sort.Strings(keys)
	return keys, nil
}

func backfillBenchmarkTxn(
	ctx context.Context,
	db client.DB,
	objCount int,
	valueSize int,
) ([]string, error) {
	clientTxn, err := db.NewTxn(ctx, false)
	if err != nil {
		return nil, err
	}
	defer clientTxn.Discard(ctx)

	txn := txnctx.MustGetFromClient(clientTxn)

	keys := make([]string, objCount)
	for i := 0; i < objCount; i++ {
		key := make([]byte, 32)
		value := make([]byte, valueSize)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		keys[i] = string(key)

		if err := txn.Rootstore().Set(ctx, key, value); err != nil {
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
	return (i * pointsPerInterval) + mathRand.Intn(pointsPerInterval)
}
