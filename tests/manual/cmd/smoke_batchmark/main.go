// smoke_batchmark exercises BatchMarkAsMerged on the P2P blockstore.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_batchmark/
package main

import (
	"context"
	"fmt"
	"os"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

func main() {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	bstore := datastore.P2PBlockstoreFrom(rootstore, immutable.None[int]())

	// Write two blocks so they have CIDs to work with.
	blk1 := blocks.NewBlock([]byte("block-one"))
	blk2 := blocks.NewBlock([]byte("block-two"))

	if err := bstore.Put(ctx, blk1); err != nil {
		fmt.Fprintln(os.Stderr, "Put blk1:", err)
		os.Exit(1)
	}
	if err := bstore.Put(ctx, blk2); err != nil {
		fmt.Fprintln(os.Stderr, "Put blk2:", err)
		os.Exit(1)
	}

	// BatchMarkAsMerged removes the to-merge entries for both CIDs.
	if err := bstore.BatchMarkAsMerged(ctx, []cid.Cid{blk1.Cid(), blk2.Cid()}); err != nil {
		fmt.Fprintln(os.Stderr, "BatchMarkAsMerged:", err)
		os.Exit(1)
	}

	fmt.Println("BatchMarkAsMerged: ok")
}
