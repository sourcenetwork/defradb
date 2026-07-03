// smoke_batchsign exercises the batch signing API:
// NewBatchCIDCollector, ContextWithBatchSigning, IsBatchSigningEnabled.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_batchsign/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/node"
)

func main() {
	ctx := context.Background()

	if node.IsBatchSigningEnabled(ctx) {
		fmt.Fprintln(os.Stderr, "FAIL: batch signing should not be enabled on a plain context")
		os.Exit(1)
	}

	collector := node.NewBatchCIDCollector()
	if collector == nil {
		fmt.Fprintln(os.Stderr, "FAIL: NewBatchCIDCollector returned nil")
		os.Exit(1)
	}

	ctx2 := node.ContextWithBatchSigning(ctx, collector)
	if !node.IsBatchSigningEnabled(ctx2) {
		fmt.Fprintln(os.Stderr, "FAIL: batch signing should be enabled after ContextWithBatchSigning")
		os.Exit(1)
	}

	// Original ctx should remain unaffected.
	if node.IsBatchSigningEnabled(ctx) {
		fmt.Fprintln(os.Stderr, "FAIL: original context should still not have batch signing")
		os.Exit(1)
	}

	fmt.Println("batch signing: ok") //nolint:forbidigo
}
