package fetcher

import (
	"testing"
	"context"

	"github.com/sourcenetwork/defradb/bench/fixtures"
	benchutils "github.com/sourcenetwork/defradb/bench"
)


func BenchmarkFilterNoFilter_1_1(b *testing.B) {
	ctx := context.Background()
	fixture := fixtures.ForSchema(ctx, "user_simple")
	db, collections, err := benchuutils.SetupDBAndCollections(b, ctx, fixture)
	

}