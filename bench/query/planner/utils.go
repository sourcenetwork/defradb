// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"
	"fmt"
	"testing"

	benchutils "github.com/sourcenetwork/defradb/bench"
	"github.com/sourcenetwork/defradb/bench/fixtures"
	"github.com/sourcenetwork/defradb/query/graphql/planner"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
)

func runQueryParserBench(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	query string,
) error {
	exec, err := buildExecutor(ctx, fixture)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("Failed to parse query string: %w", err)
		}
	}
	b.StopTimer()

	return nil
}

func runMakePlanBench(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	query string,
) error {
	db, _, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer db.Close(ctx)

	exec, err := buildExecutor(ctx, fixture)
	if err != nil {
		return err
	}

	q, err := exec.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("Failed to parse query string: %w", err)
	}
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

func buildExecutor(
	ctx context.Context,
	fixture fixtures.Generator,
) (*planner.QueryExecutor, error) {
	sm, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}
	schema, err := benchutils.ConstructSchema(fixture)
	if err != nil {
		return nil, err
	}
	types, _, err := sm.Generator.FromSDL(ctx, schema)
	if err != nil {
		return nil, err
	}
	_, err = sm.Generator.CreateDescriptions(types)
	if err != nil {
		return nil, err
	}

	return planner.NewQueryExecutor(sm)
}
