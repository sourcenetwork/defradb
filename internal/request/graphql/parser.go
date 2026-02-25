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
	// mu protects schemaManager. Writers (SetSchema's OnSuccess callback) hold
	// the full write lock while swapping the pointer; readers hold a shared read
	// lock only for the duration of the pointer dereference + snapshot copy.
	// This keeps the critical section extremely short and avoids blocking normal
	// GQL request execution during schema mutations.
	mu                            sync.RWMutex
	schemaManager                 *schema.SchemaManager
	isSearchableEncryptionEnabled bool
}

func NewParser(isSearchableEncryptionEnabled bool) (*parser, error) {
	schemaManager, err := schema.NewSchemaManager(isSearchableEncryptionEnabled)
	if err != nil {
		return nil, err
	}

	p := &parser{
		schemaManager:                 schemaManager,
		isSearchableEncryptionEnabled: isSearchableEncryptionEnabled,
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
	p.mu.RLock()
	schema := p.schemaManager.Schema()
	p.mu.RUnlock()
	return defrap.IsIntrospectionQuery(*schema, ast)
}

func (p *parser) ExecuteIntrospection(ctx context.Context, request string) *client.RequestResult {
	_, span := tracer.Start(ctx)
	defer span.End()

	p.mu.RLock()
	schema := p.schemaManager.Schema()
	p.mu.RUnlock()

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

	// Snapshot the schema pointer under a short-lived read lock. The schema object
	// itself is immutable once published, so it is safe to use it after releasing
	// the lock.
	p.mu.RLock()
	schema := p.schemaManager.Schema()
	p.mu.RUnlock()

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

	p.mu.RLock()
	sm := p.schemaManager
	p.mu.RUnlock()

	return sm.ParseSDL(sdl)
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

	txn := datastore.CtxMustGetTxn(ctx)

	txn.OnSuccess(
		func() {
			// The write lock is held only for the pointer swap, keeping the critical
			// section as short as possible. Readers (Parse, IsIntrospection, etc.)
			// snapshot the pointer under a shared read lock and then use the
			// immutable schema object without holding any lock.
			p.mu.Lock()
			p.schemaManager = schemaManager
			p.mu.Unlock()
		},
	)
	return err
}

func (p *parser) NewFilterFromString(collectionType string, body string) (immutable.Option[request.Filter], error) {
	p.mu.RLock()
	schema := p.schemaManager.Schema()
	p.mu.RUnlock()
	return defrap.NewFilterFromString(*schema, collectionType, body)
}
