// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"context"
	"sync"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	defrap "github.com/sourcenetwork/defradb/internal/request/graphql/parser"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
	"github.com/sourcenetwork/defradb/internal/telemetry"
)

var _ core.Parser = (*parser)(nil)

var tracer = telemetry.NewTracer()

type parser struct {
	schemaManager                 *schema.SchemaManager
	isSearchableEncryptionEnabled bool
	// In the cases of transactions, we need to store a schema manager for each transaction
	schemaManagerMapLock sync.RWMutex
	schemaManagerMap     map[uint64]*schema.SchemaManager
}

func NewParser(isSearchableEncryptionEnabled bool) (*parser, error) {
	schemaManager, err := schema.NewSchemaManager(isSearchableEncryptionEnabled)
	if err != nil {
		return nil, err
	}

	p := &parser{
		schemaManager:                 schemaManager,
		isSearchableEncryptionEnabled: isSearchableEncryptionEnabled,
		schemaManagerMapLock:          sync.RWMutex{},
		schemaManagerMap:              make(map[uint64]*schema.SchemaManager),
	}

	return p, nil
}

func (p *parser) BuildRequestAST(ctx context.Context, request string) (*ast.Document, error) {
	_, span := tracer.Start(ctx)
	defer span.End()

	source := source.NewSource(&source.Source{
		Body: []byte(request),
		Name: "GraphQL request",
	})

	ast, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func (p *parser) IsIntrospection(ast *ast.Document) bool {
	schema := p.schemaManager.Schema()
	return defrap.IsIntrospectionQuery(*schema, ast)
}

func (p *parser) ExecuteIntrospection(ctx context.Context, request string) *client.RequestResult {
	_, span := tracer.Start(ctx)
	defer span.End()

	schema := p.schemaManager.Schema()
	params := gql.Params{Schema: *schema, RequestString: request}
	r := gql.Do(params)

	res := &client.RequestResult{
		GQL: client.GQLResult{
			Data: r.Data,
		},
	}

	for _, err := range r.Errors {
		res.GQL.Errors = append(res.GQL.Errors, err)
	}

	return res
}

func (p *parser) Parse(ctx context.Context, ast *ast.Document, options *client.GQLOptions) (*request.Request, []error) {
	_, span := tracer.Start(ctx)
	defer span.End()

	// If there is a transaction, we will check to see if we have a store schema manager for it
	// If we don't, or if we don't have a transaction at all, then we use the default schema manager
	gotTxn, hadTxn := datastore.CtxTryGetTxn(ctx)
	schema := p.schemaManager.Schema()
	if hadTxn {
		p.schemaManagerMapLock.RLock()
		gotSchemaManager, ok := p.schemaManagerMap[gotTxn.ID()]
		p.schemaManagerMapLock.RUnlock()
		if ok {
			schema = gotSchemaManager.Schema()
		} else {
			schema = p.schemaManager.Schema()
		}
	}

	validationResult := gql.ValidateDocument(schema, ast, nil)
	if !validationResult.IsValid {
		errors := make([]error, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			errors[i] = err
		}
		return nil, errors
	}

	return defrap.ParseRequest(*schema, ast, options)
}

func (p *parser) ParseSDL(ctx context.Context, sdl string) ([]core.Collection, error) {
	_, span := tracer.Start(ctx)
	defer span.End()

	return p.schemaManager.ParseSDL(sdl)
}

func (p *parser) SetSchema(ctx context.Context, collections []client.CollectionVersion) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	schemaManager, err := schema.NewSchemaManager(p.isSearchableEncryptionEnabled)
	if err != nil {
		return err
	}

	_, err = schemaManager.Generator.Generate(ctx, collections)
	if err != nil {
		return err
	}

	// If we had a transaction, map its transaction ID to a schema manager unique to it
	gotTxn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if hadTxn {
		p.schemaManagerMapLock.Lock()
		p.schemaManagerMap[gotTxn.ID()] = schemaManager
		p.schemaManagerMapLock.Unlock()
	}

	txn := datastore.CtxMustGetTxn(ctx)

	txn.OnSuccess(
		func() {
			p.schemaManager = schemaManager
			// If the txn ID is in the schema manager map, remove it
			p.schemaManagerMapLock.Lock()
			delete(p.schemaManagerMap, txn.ID())
			p.schemaManagerMapLock.Unlock()
		},
	)

	txn.OnDiscard(
		func() {
			// If the txn ID is in the schema manager map, remove it
			p.schemaManagerMapLock.Lock()
			delete(p.schemaManagerMap, txn.ID())
			p.schemaManagerMapLock.Unlock()
		},
	)
	return err
}

func (p *parser) NewFilterFromString(collectionType string, body string) (immutable.Option[request.Filter], error) {
	return defrap.NewFilterFromString(*p.schemaManager.Schema(), collectionType, body)
}
