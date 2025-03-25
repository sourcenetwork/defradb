package main

import (
	"context"

	"github.com/sourcenetwork/defradb"

	"github.com/sourcenetwork/corekv/memory"
)

func main() {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := defradb.Open(ctx, store)
	if err != nil {
		return
	}

	db.Close()
}
