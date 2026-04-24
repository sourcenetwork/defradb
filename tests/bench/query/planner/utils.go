// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package planner

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/request/graphql"
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
		ast, _ := parser.BuildRequestAST(ctx, query)
		_, errs := parser.Parse(ctx, ast, &client.GQLOptions{})
		if errs != nil {
			return errors.Wrap("failed to parse query string", errors.New(fmt.Sprintf("%v", errs)))
		}
	}
	b.StopTimer()

	return nil
}

func buildParser(
	ctx context.Context,
	fixture fixtures.Generator,
) (core.Parser, error) {
	sdl, err := benchutils.ConstructSDL(fixture)
	if err != nil {
		return nil, err
	}

	parser, err := graphql.NewParser(false)
	if err != nil {
		return nil, err
	}

	collectionVersions, err := parser.ParseSDL(ctx, sdl)
	if err != nil {
		return nil, err
	}

	collections := make([]client.CollectionVersion, len(collectionVersions))
	for i, collectionVersion := range collectionVersions {
		collections[i] = collectionVersion.Definition
	}

	err = parser.SetSchema(ctx, collections)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
