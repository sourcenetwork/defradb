package storage

import (
	"context"
	"math/rand"
	"testing"

	ds "github.com/ipfs/go-datastore"

	testutils "github.com/sourcenetwork/defradb/db/tests"
)

func runStorageBenchGet(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := testutils.NewTestStorage(b)
	if err != nil {
		return err
	}
	defer db.Close() // @todo: File based needs to handle proper temp file cleanup

	// backfill
	keys, err := backfillBenchmarkStorageDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	//shuffle keys
	// rand.Shuffle(len(keys), func(i, j int) {
	// 	keys[i], keys[j] = keys[j], keys[i]
	// })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			key := ds.NewKey(keys[rand.Int31n(int32(len(keys)))])
			_, err := db.Get(ctx, key)
			if err != nil {
				return err
			}
		}
	}
	b.StopTimer()

	return nil
}

func runStorageBenchPut(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := testutils.NewTestStorage(b)
	if err != nil {
		return err
	}
	defer db.Close() // @todo: File based needs to handle proper temp file cleanup

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
	db, err := testutils.NewTestStorage(b)
	if err != nil {
		return err
	}
	defer db.Close() // @todo: File based needs to handle proper temp file cleanup

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
		keyBuf := make([]byte, 32)
		value := make([]byte, valueSize)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		if _, err := rand.Read(keyBuf); err != nil {
			return nil, err
		}
		key := ds.NewKey(string(keyBuf))
		keys[i] = key.String()

		if err := batch.Put(ctx, key, value); err != nil {
			return nil, err
		}
	}

	return keys, batch.Commit(ctx)
}
