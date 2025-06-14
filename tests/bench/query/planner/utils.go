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

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner"
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

func runMakePlanBench(
	b *testing.B,
	ctx context.Context,
	fixture fixtures.Generator,
	query string,
) error {
	d, _, err := benchutils.SetupDBAndCollections(b, ctx, fixture)
	if err != nil {
		return err
	}
	defer d.Close()

	parser, err := buildParser(ctx, fixture)
	if err != nil {
		return err
	}

	ast, _ := parser.BuildRequestAST(ctx, query)
	q, errs := parser.Parse(ctx, ast, &client.GQLOptions{})
	if len(errs) > 0 {
		return errors.Wrap("failed to parse query string", errors.New(fmt.Sprintf("%v", errs)))
	}
	txn, err := d.NewTxn(ctx, false)
	if err != nil {
		return errors.Wrap("failed to create txn", err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		planner := planner.New(
			ctx,
			acpIdentity.None,
			dac.NoDocumentACP,
			d,
			txn,
		)
		plan, err := planner.MakePlan(q)
		if err != nil {
			return errors.Wrap("failed to make plan", err)
		}

		err = plan.Init()
		if err != nil {
			return errors.Wrap("failed to init plan", err)
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

	collectionVersions, err := parser.ParseSDL(ctx, schema)
	if err != nil {
		return nil, err
	}

	collectionDefinitions := make([]client.CollectionDefinition, len(collectionVersions))
	for i, collectionVersion := range collectionVersions {
		collectionDefinitions[i] = collectionVersion.Definition
	}

	err = parser.SetSchema(ctx, &dummyTxn{}, collectionDefinitions)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

var _ datastore.Txn = (*dummyTxn)(nil)

type dummyTxn struct{}

func (*dummyTxn) Rootstore() datastore.DSReaderWriter   { return nil }
func (*dummyTxn) Datastore() datastore.DSReaderWriter   { return nil }
func (*dummyTxn) Encstore() datastore.Blockstore        { return nil }
func (*dummyTxn) Headstore() datastore.DSReaderWriter   { return nil }
func (*dummyTxn) Peerstore() datastore.DSReaderWriter   { return nil }
func (*dummyTxn) Blockstore() datastore.Blockstore      { return nil }
func (*dummyTxn) Systemstore() datastore.DSReaderWriter { return nil }
func (*dummyTxn) Commit(ctx context.Context) error      { return nil }
func (*dummyTxn) Discard(ctx context.Context)           {}
func (*dummyTxn) OnSuccess(fn func())                   {}
func (*dummyTxn) OnError(fn func())                     {}
func (*dummyTxn) OnDiscard(fn func())                   {}
func (*dummyTxn) OnSuccessAsync(fn func())              {}
func (*dummyTxn) OnErrorAsync(fn func())                {}
func (*dummyTxn) OnDiscardAsync(fn func())              {}
func (*dummyTxn) ID() uint64                            { return 0 }
