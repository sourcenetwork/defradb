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

	// Checked before validation because the generic "unknown argument" error the library would
	// otherwise produce hides what is actually wrong.
	if errs := validateSimilarityArgs(schema, ast); len(errs) > 0 {
		return nil, errs
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

func (p *parser) NewFilterFromString(
	ctx context.Context,
	collectionType string,
	body string) (immutable.Option[request.Filter], error) {
	// If there was a transaction, try to use the schema manager for that transaction
	txn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if hadTxn {
		p.schemaManagerMapLock.RLock()
		gotSchemaManager, ok := p.schemaManagerMap[txn.ID()]
		p.schemaManagerMapLock.RUnlock()
		if ok {
			return defrap.NewFilterFromString(*gotSchemaManager.Schema(), collectionType, body)
		}
	}
	return defrap.NewFilterFromString(*p.schemaManager.Schema(), collectionType, body)
}

// validateSimilarityArgs reports SIMILARITY arguments naming a field that exists but is not a
// numeric array. SIMILARITY only gets an argument per numeric-array field, so the library reports
// those as unknown arguments, which reads as if the field itself does not exist.
func validateSimilarityArgs(schema *gql.Schema, doc *ast.Document) []error {
	fragments := map[string]*ast.FragmentDefinition{}
	for _, definition := range doc.Definitions {
		if fragment, isFragment := definition.(*ast.FragmentDefinition); isFragment {
			fragments[fragment.Name.Value] = fragment
		}
	}

	var errs []error
	for _, definition := range doc.Definitions {
		operation, isOperation := definition.(*ast.OperationDefinition)
		if !isOperation {
			continue
		}
		// SIMILARITY is generated on every object type, so it can be selected from a mutation's
		// result set as readily as from a query.
		root := schema.QueryType()
		switch operation.Operation {
		case ast.OperationTypeMutation:
			root = schema.MutationType()
		case ast.OperationTypeSubscription:
			root = schema.SubscriptionType()
		}
		if root == nil {
			continue
		}
		errs = append(errs, checkSimilarityArgs(root, operation.SelectionSet, fragments, map[string]bool{})...)
	}
	return errs
}

// checkSimilarityArgs checks the SIMILARITY selections made on obj, then recurses into the related
// objects selected alongside them. visited guards against a fragment cycle.
func checkSimilarityArgs(
	obj *gql.Object,
	selectionSet *ast.SelectionSet,
	fragments map[string]*ast.FragmentDefinition,
	visited map[string]bool,
) []error {
	if obj == nil || selectionSet == nil {
		return nil
	}

	similarityArgs := map[string]struct{}{}
	if similarity, exists := obj.Fields()[request.SimilarityFieldName]; exists {
		for _, arg := range similarity.Args {
			similarityArgs[arg.Name()] = struct{}{}
		}
	}

	var errs []error
	for _, selection := range selectionSet.Selections {
		switch node := selection.(type) {
		case *ast.InlineFragment:
			errs = append(errs, checkSimilarityArgs(obj, node.SelectionSet, fragments, visited)...)

		case *ast.FragmentSpread:
			name := node.Name.Value
			if visited[name] {
				continue
			}
			visited[name] = true
			if fragment, exists := fragments[name]; exists {
				errs = append(errs, checkSimilarityArgs(obj, fragment.SelectionSet, fragments, visited)...)
			}

		case *ast.Field:
			if node.Name.Value != request.SimilarityFieldName {
				errs = append(errs, checkSimilarityArgs(
					objectOf(obj, node.Name.Value), node.SelectionSet, fragments, visited)...)
				continue
			}
			errs = append(errs, checkSimilarityFieldArgs(obj, node, similarityArgs)...)
		}
	}
	return errs
}

// checkSimilarityFieldArgs checks one SIMILARITY selection's arguments against obj's fields.
func checkSimilarityFieldArgs(
	obj *gql.Object,
	similarity *ast.Field,
	similarityArgs map[string]struct{},
) []error {
	var errs []error
	for _, arg := range similarity.Arguments {
		name := arg.Name.Value
		if _, isSimilarityArg := similarityArgs[name]; isSimilarityArg {
			continue
		}
		field, exists := obj.Fields()[name]
		if !exists {
			// Not a field at all, so the library's "unknown argument" error is already right.
			continue
		}
		errs = append(errs, NewErrSimilarityOnNonVectorField(name, field.Type.String()))
	}
	return errs
}

// objectOf resolves the object type behind the named field of obj, looking through the list and
// non-null wrappers a collection or relation field is built from.
func objectOf(obj *gql.Object, fieldName string) *gql.Object {
	field, exists := obj.Fields()[fieldName]
	if !exists {
		return nil
	}

	typ := field.Type
	for {
		switch unwrapped := typ.(type) {
		case *gql.List:
			typ = unwrapped.OfType
		case *gql.NonNull:
			typ = unwrapped.OfType
		case *gql.Object:
			return unwrapped
		default:
			return nil
		}
	}
}
