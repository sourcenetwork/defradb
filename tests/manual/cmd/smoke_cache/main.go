// smoke_cache exercises MergedCIDCache: Add, Has, FIFO eviction, and BatchAdd.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_cache/
package main

import (
	"fmt"
	"os"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

func makeCID(seed string) cid.Cid {
	hash, err := mh.Sum([]byte(seed), mh.SHA2_256, -1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mh.Sum:", err)
		os.Exit(1)
	}
	return cid.NewCidV1(cid.Raw, hash)
}

func main() {
	const capacity = 4
	cache := coreblock.NewMergedCIDCache(capacity)

	first := makeCID("first")
	if cache.Has(first) {
		fmt.Fprintln(os.Stderr, "FAIL: fresh cache should not contain 'first'")
		os.Exit(1)
	}

	cache.Add(first)
	if !cache.Has(first) {
		fmt.Fprintln(os.Stderr, "FAIL: 'first' should be cached after Add")
		os.Exit(1)
	}

	// Fill beyond capacity to evict 'first' via FIFO.
	for i := 0; i <= capacity; i++ {
		cache.Add(makeCID(fmt.Sprintf("extra-%d", i)))
	}
	if cache.Has(first) {
		fmt.Fprintln(os.Stderr, "FAIL: 'first' should have been evicted")
		os.Exit(1)
	}

	// BatchAdd
	batch := []cid.Cid{makeCID("b1"), makeCID("b2"), makeCID("b3")}
	cache.BatchAdd(batch)
	for _, c := range batch {
		if !cache.Has(c) {
			fmt.Fprintf(os.Stderr, "FAIL: BatchAdd entry %s missing\n", c)
			os.Exit(1)
		}
	}

	fmt.Println("MergedCIDCache: ok")
}
