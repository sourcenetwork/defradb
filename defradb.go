package defradb

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
)

func Open(ctx context.Context, rootstore datastore.Rootstore) (client.DB, error) {
	return db.NewDB(ctx, rootstore)
}
