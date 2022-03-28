package badger

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	badger "github.com/dgraph-io/badger/v3"
)

var (
	storage = "badger"
)

func runBadgerBenchGet(b *testing.B, ctx context.Context, valueSize, objCount, opCount int, doSync bool) error {
	db, err := newBadgerDB(b)
	if err != nil {
		return err
	}
	defer db.Close() //nolint

	// backfill
	keys, err := backfillBenchmarkBadgerDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < opCount; j++ {
			positionInInterval := getSampledIndex(len(keys), opCount, j)
			key := keys[positionInInterval]
			err := db.View(func(txn *badger.Txn) error {
				_, err := txn.Get([]byte(key))
				return err
			})
			if err != nil {
				return err
			}

		}
	}
	b.StopTimer()

	return nil
}

func runBadgerIteratorKeysOnly(b *testing.B, ctx context.Context, valueSize int, objCount int) error {
	db, err := newBadgerDB(b)
	if err != nil {
		return err
	}
	defer db.Close()

	// backfill
	_, err = backfillBenchmarkBadgerDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		txn := db.NewTransaction(false)

		// iterate over all keys
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		b.StartTimer()

		resCount := 0
		for it.Rewind(); it.Valid(); it.Next() {
			resCount++
			item := it.Item()
			_ = item.Key()
			// _, err = item.ValueCopy(nil)
			// if err != nil {
			// 	return err
			// }
		}
		if resCount != objCount {
			return fmt.Errorf("incorrect query iterator doc count, expected %v got %v", objCount, resCount)
		}

		b.StopTimer()
		it.Close()
		txn.Discard()
	}

	return nil
}

func runBadgerIteratorWithValues(b *testing.B, ctx context.Context, valueSize int, objCount int, prefetch bool) error {
	db, err := newBadgerDB(b)
	if err != nil {
		return err
	}
	defer db.Close()

	// backfill
	_, err = backfillBenchmarkBadgerDB(ctx, db, objCount, valueSize)
	if err != nil {
		return err
	}

	if err := db.Sync(); err != nil {
		return err
	}

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		txn := db.NewTransaction(false)

		// iterate over all keys
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = prefetch
		it := txn.NewIterator(opts)
		b.StartTimer()

		resCount := 0
		totalBytesRead := 0
		for it.Rewind(); it.Valid(); it.Next() {
			resCount++
			item := it.Item()
			_ = item.Key()
			bz, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			totalBytesRead += len(bz) // make sure the compiler doesn't do some
			// weird optimization and silently ignore
		}
		if resCount != objCount {
			return fmt.Errorf("incorrect query iterator doc count, expected %v got %v", objCount, resCount)
		}

		if totalBytesRead != valueSize*objCount {
			return fmt.Errorf("incorrect total bytes read, expected %v got %v", valueSize*objCount, totalBytesRead)
		}

		// fmt.Println("Records:", resCount, "Bytes Read:", totalSize)

		b.StopTimer()
		it.Close()
		txn.Discard()
	}

	return nil
}

func backfillBenchmarkBadgerDB(ctx context.Context, db *badger.DB, objCount int, valueSize int) ([]string, error) {
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
		key := "/data" + string(keyBuf)
		keys[i] = key

		if err := db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), value)
		}); err != nil {
			return nil, err
		}
	}

	sort.Strings(keys)
	return keys, nil
}

// func backfillBenchmarkTxn(ctx context.Context, db *badger.DB, objCount int, valueSize int) ([]string, error) {
// 	txn, err := db.NewTxn(ctx, false)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer txn.Discard(ctx)

// 	keys := make([]string, objCount)
// 	for i := 0; i < objCount; i++ {
// 		// keyBuf := make([]byte, 32)
// 		// value := make([]byte, valueSize)
// 		// if _, err := rand.Read(value); err != nil {
// 		// 	return nil, err
// 		// }
// 		// if _, err := rand.Read(keyBuf); err != nil {
// 		// 	return nil, err
// 		// }
// 		keyBuf := randSeq(32)
// 		value := []byte(randSeq(valueSize))
// 		key := ds.NewKey("/data" + string(keyBuf))
// 		keys[i] = key.String()

// 		if err := txn.Rootstore().Put(ctx, key, value); err != nil {
// 			return nil, err
// 		}
// 	}

// 	sort.Strings(keys)
// 	return keys, txn.Commit(ctx)
// }

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

func newBadgerDB(b *testing.B, s ...string) (*badger.DB, error) {
	var store string
	if len(s) > 0 {
		store = s[0]
	} else {
		store = storage
	}
	switch store {
	case "memory":
		return newMemoryDB()
	case "badger":
		// fmt.Println("OPENING BADGER FILE DB")
		return newFileDB(b)
	}

	return nil, errors.New("Invalid DB storage option")
}

func newMemoryDB() (*badger.DB, error) {
	opts := badger.DefaultOptions("").WithInMemory(true)
	opts.SyncWrites = true
	opts.Logger = nil // badger is too chatty by default
	return badger.Open(opts)
}

func newFileDB(b *testing.B) (*badger.DB, error) {
	path := b.TempDir() + "1"
	// fmt.Println("PATH:", path)
	opts := badger.DefaultOptions(path)
	// opts.SyncWrites = true
	opts.Logger = nil
	return badger.Open(opts)
}
