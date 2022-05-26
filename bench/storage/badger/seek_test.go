package badger

import (
	"context"
	"fmt"
	"testing"
)

// func TestBadgerSeekDirections(t *testing.T) {
// 	db, err := newBadgerDB(t)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer db.Close() //nolint

// 	ctx := context.Background()

// 	// backfill
// 	keys, err := backfillBenchmarkBadgerDB(ctx, db, 10, 64)
// 	for _, k := range keys {
// 		fmt.Println(k)
// 	}

// 	fmt.Println("---------------")
// 	txn := db.NewTransaction(false)
// 	// iterate over all keys
// 	opts := badger.DefaultIteratorOptions
// 	opts.PrefetchValues = false
// 	it := txn.NewIterator(opts)

// 	it.Rewind()
// 	for i := 0; i < 5 && it.Valid(); it.Next() {
// 		item := it.Item()
// 		key := item.Key()
// 		if string(key) != keys[i] {
// 			t.Fatal(fmt.Errorf("wrong keys expected %v actual %v", keys[i], string(key)))
// 		}

// 		fmt.Println(string(key))
// 		i++
// 	}

// 	fmt.Println("--------------")

// 	// back to the beginning
// 	it.Seek([]byte(keys[0]))

// 	for i := 0; i < 5 && it.Valid(); it.Next() {
// 		item := it.Item()
// 		key := item.Key()
// 		if string(key) != keys[i] {
// 			t.Fatal(fmt.Errorf("wrong keys expected %v actual %v", keys[i], string(key)))
// 		}

// 		fmt.Println(string(key))
// 		i++
// 	}
// }

func Benchmark_Badger_Simple_Seek_WithoutPrefetch_1000_500_500(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerIteratorSeek(b, ctx, vsz, 1000, 500, 500, false)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Seek_WithoutPrefetch_1000_1000_500(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerIteratorSeek(b, ctx, vsz, 1000, 1000, 500, false)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Seek_WithoutPrefetch_1000_1_1(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerIteratorSeek(b, ctx, vsz, 1000, 1000, 500, false)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Seek_WithPrefetch_1000_500_500(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerIteratorSeek(b, ctx, vsz, 1000, 500, 500, true)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}

func Benchmark_Badger_Simple_Seek2_WithoutPrefetch_100000(b *testing.B) {
	for _, vsz := range valueSize {
		b.Run(fmt.Sprintf("ValueSize:%04d", vsz), func(b *testing.B) {
			ctx := context.Background()
			err := runBadgerIteratorSeek2(b, ctx, vsz, 100000, false)
			if err != nil {
				b.Fatal(err)
			}
		})
	}
}
