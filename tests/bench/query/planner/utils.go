// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/planner"
	"github.com/sourcenetwork/defradb/request/graphql"
	benchutils "github.com/sourcenetwork/defradb/tests/bench"
	"github.com/sourcenetwork/defradb/tests/bench/fixtures"
)

func runQueryParserBench(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	query string,
) error {
	parser, err := buildParser(ctx, fixture)
	if err != nil {
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, errs := parser.Parse(query)
		if errs != nil {
			return errors.Wrap("failed to parse query string", errors.New(fmt.Sprintf("%v", errs)))
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

	parser, err := buildParser(ctx, fixture)
	if err != nil {
		return err
	}

	q, errs := parser.Parse(query)
	if len(errs) > 0 {
		return errors.Wrap("failed to parse query string", errors.New(fmt.Sprintf("%v", errs)))
	}
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return errors.Wrap("failed to create txn", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		planner := planner.New(ctx, db, txn)
		_, err := planner.MakePlan(q)
		if err != nil {
			return errors.Wrap("failed to make plan", err)
		}
	}
	b.StopTimer()
	return nil
}

func buildParser(
	ctx context.Context,
	fixture fixtures.Generator,
) (core.Parser, error) {
	schema, err := benchutils.ConstructSchema(fixture)
	if err != nil {
		return nil, err
	}

	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}
	err = parser.AddSchema(ctx, schema)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
