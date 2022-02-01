package planner

import (
	"context"
	"fmt"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
)

func runQueryParserBench(b *testing.B, ctx context.Context, fixture fixtures.Generator, query string) error {
	db, _, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Executor().ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("Failed to parse query string: %w", err)
		}
	}
	b.StopTimer()

	return nil
}

func runMakePlanBench(b *testing.B, ctx context.Context, fixture fixtures.Generator, query string) error {
	db, _, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close()

	exec := db.Executor()
	if exec == nil {
		return fmt.Errorf("Executor can't be nil")
	}

	q, err := exec.ParseQueryString(query)
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return fmt.Errorf("Failed to create txn: %w", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.MakePlanFromParser(ctx, db, txn, q)
		if err != nil {
			return fmt.Errorf("Failed to make plan: %w", err)
		}
	}
	b.StopTimer()
	return nil
}
