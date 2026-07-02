// smoke_bbs exercises the blind-write blockstore: verifies that Put writes the
// block without a Has() pre-check and that the block is subsequently readable.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_bbs/
package main

import (
	"context"
	"fmt"
	"os"

	blocks "github.com/ipfs/go-block-format"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

func main() {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	bstore := datastore.BlindWriteP2PBlockstoreFrom(rootstore, immutable.None[int]())

	blk := blocks.NewBlock([]byte("hello blind write"))

	if err := bstore.Put(ctx, blk); err != nil {
		fmt.Fprintln(os.Stderr, "Put:", err)
		os.Exit(1)
	}

	ok, err := bstore.Has(ctx, blk.Cid())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Has:", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "FAIL: block not found after Put")
		os.Exit(1)
	}

	got, err := bstore.Get(ctx, blk.Cid())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Get:", err)
		os.Exit(1)
	}
	if string(got.RawData()) != "hello blind write" {
		fmt.Fprintln(os.Stderr, "FAIL: wrong block data:", string(got.RawData()))
		os.Exit(1)
	}

	fmt.Println("blind-write blockstore: ok")
}
