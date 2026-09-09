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
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/sourcenetwork/immutable"
	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/variablesvalidation"

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

type requestAST struct {
	tools  wgast.Document
	source string
}

func (*requestAST) Language() string {
	return "graphql"
}

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

func (p *parser) BuildRequestAST(ctx context.Context, request string) (core.RequestAST, error) {
	_, span := tracer.Start(ctx)
	defer span.End()

	toolsDocument, toolsReport := astparser.ParseGraphqlDocumentString(request)
	if toolsReport.HasErrors() {
		return nil, fmt.Errorf("Syntax Error GraphQL: %w", toolsReport)
	}

	return &requestAST{tools: toolsDocument, source: request}, nil
}

func (p *parser) IsIntrospection(document core.RequestAST) bool {
	ast, ok := document.(*requestAST)
	if !ok {
		return false
	}
	for _, operation := range ast.tools.OperationDefinitions {
		if operation.OperationType != wgast.OperationTypeQuery {
			continue
		}
		if selectionSetContainsIntrospection(&ast.tools, operation.SelectionSet, map[int]bool{}) {
			return true
		}
	}
	return false
}

func selectionSetContainsIntrospection(document *wgast.Document, selectionSet int, visited map[int]bool) bool {
	for _, selectionRef := range document.SelectionSets[selectionSet].SelectionRefs {
		selection := document.Selections[selectionRef]
		switch selection.Kind {
		case wgast.SelectionKindField:
			name := document.FieldNameString(selection.Ref)
			if name == "__schema" || name == "__type" {
				return true
			}
		case wgast.SelectionKindInlineFragment:
			fragment := document.InlineFragments[selection.Ref]
			if fragment.HasSelections && selectionSetContainsIntrospection(document, fragment.SelectionSet, visited) {
				return true
			}
		case wgast.SelectionKindFragmentSpread:
			fragmentRef, ok := document.FragmentDefinitionRef(document.FragmentSpreadNameBytes(selection.Ref))
			if !ok || visited[fragmentRef] {
				continue
			}
			visited[fragmentRef] = true
			if selectionSetContainsIntrospection(document, document.FragmentDefinitions[fragmentRef].SelectionSet, visited) {
				return true
			}
		}
	}
	return false
}

func (p *parser) ExecuteIntrospection(ctx context.Context, request string) *client.RequestResult {
	_, span := tracer.Start(ctx)
	defer span.End()

	return executeIntrospection(p.schemaManager.Definition(), request)
}

func (p *parser) Parse(
	ctx context.Context,
	document core.RequestAST,
	options *client.GQLOptions,
) (*request.Request, []error) {
	_, span := tracer.Start(ctx)
	defer span.End()
	ast, ok := document.(*requestAST)
	if !ok {
		return nil, []error{fmt.Errorf("unexpected request AST type %T", document)}
	}

	// If there is a transaction, we will check to see if we have a store schema manager for it
	// If we don't, or if we don't have a transaction at all, then we use the default schema manager
	gotTxn, hadTxn := datastore.CtxTryGetTxn(ctx)
	schemaManager := p.schemaManager
	if hadTxn {
		p.schemaManagerMapLock.RLock()
		gotSchemaManager, ok := p.schemaManagerMap[gotTxn.ID()]
		p.schemaManagerMapLock.RUnlock()
		if ok {
			schemaManager = gotSchemaManager
		}
	}
	toolsReport := operationreport.Report{}
	astnormalization.NormalizeOperation(&ast.tools, schemaManager.Definition(), &toolsReport)
	if toolsReport.HasErrors() {
		errs := append([]error(nil), toolsReport.InternalErrors...)
		for _, gqlErr := range toolsReport.ExternalErrors {
			errs = append(errs, fmt.Errorf("%s", gqlErr.Message))
		}
		return nil, errs
	}
	if errs := defrap.ValidateSimilarityArgs(schemaManager.Definition(), &ast.tools); len(errs) > 0 {
		return nil, errs
	}
	if astvalidation.DefaultOperationValidator().Validate(
		&ast.tools,
		schemaManager.Definition(),
		&toolsReport,
	) == astvalidation.Invalid {
		errs := append([]error(nil), toolsReport.InternalErrors...)
		for _, gqlErr := range toolsReport.ExternalErrors {
			errs = append(errs, fmt.Errorf("%s", gqlErr.Message))
		}
		return nil, errs
	}
	selectedOperation, err := selectAndNormalizeOperation(
		ast.source,
		options.OperationName,
		options.Variables,
		schemaManager.Definition(),
	)
	if err != nil {
		return nil, []error{err}
	}
	if err := validateVariables(&selectedOperation, options, schemaManager.Definition()); err != nil {
		return nil, []error{err}
	}

	return defrap.ParseRequest(schemaManager.Definition(), &selectedOperation)
}

func selectAndNormalizeOperation(
	source string,
	operationName string,
	variables map[string]any,
	definition *wgast.Document,
) (wgast.Document, error) {
	operation, report := astparser.ParseGraphqlDocumentString(source)
	if report.HasErrors() {
		return wgast.Document{}, report
	}

	switch {
	case len(operation.OperationDefinitions) == 0:
		return wgast.Document{}, fmt.Errorf("Must provide an operation.")
	case operationName == "" && len(operation.OperationDefinitions) > 1:
		return wgast.Document{}, fmt.Errorf("Must provide operation name if query contains multiple operations.")
	case operationName == "":
		operationName = operation.OperationDefinitionNameString(0)
	default:
		found := false
		for ref := range operation.OperationDefinitions {
			if operation.OperationDefinitionNameString(ref) == operationName {
				found = true
				break
			}
		}
		if !found {
			return wgast.Document{}, fmt.Errorf("Unknown operation named %q.", operationName)
		}
	}

	variablesWithDefaults := make(map[string]any, len(variables))
	for name, value := range variables {
		variablesWithDefaults[name] = value
	}
	for operationRef := range operation.OperationDefinitions {
		if operation.OperationDefinitionNameString(operationRef) != operationName && operationName != "" {
			continue
		}
		for _, variableRef := range operation.OperationDefinitions[operationRef].VariableDefinitions.Refs {
			name := operation.VariableDefinitionNameString(variableRef)
			if _, exists := variablesWithDefaults[name]; exists || !operation.VariableDefinitionHasDefaultValue(variableRef) {
				continue
			}
			encodedDefault, err := operation.ValueToJSON(operation.VariableDefinitionDefaultValue(variableRef))
			if err != nil {
				return wgast.Document{}, err
			}
			var defaultValue any
			if err := json.Unmarshal(encodedDefault, &defaultValue); err != nil {
				return wgast.Document{}, err
			}
			variablesWithDefaults[name] = defaultValue
		}
	}
	encodedVariables, err := json.Marshal(variablesWithDefaults)
	if err != nil {
		return wgast.Document{}, err
	}
	operation.Input.Variables = encodedVariables

	normalizer := astnormalization.NewWithOpts(
		astnormalization.WithRemoveNotMatchingOperationDefinitions(),
		astnormalization.WithRemoveFragmentDefinitions(),
		astnormalization.WithInlineFragmentSpreads(),
		astnormalization.WithRemoveUnusedVariables(),
	)
	normalizer.NormalizeNamedOperation(&operation, definition, []byte(operationName), &report)
	if report.HasErrors() {
		return wgast.Document{}, report
	}
	return operation, nil
}

func validateVariables(operation *wgast.Document, options *client.GQLOptions, definition *wgast.Document) error {
	variables := []byte("{}")
	if options.Variables != nil {
		coerced := make(map[string]any, len(options.Variables))
		for name, value := range options.Variables {
			coerced[name] = value
		}
		for _, root := range operation.RootNodes {
			if root.Kind != wgast.NodeKindOperationDefinition {
				continue
			}
			for _, ref := range operation.OperationDefinitions[root.Ref].VariableDefinitions.Refs {
				name := operation.VariableDefinitionNameString(ref)
				value, exists := coerced[name]
				if !exists || value == nil {
					continue
				}
				typeRef := operation.VariableDefinitionType(ref)
				if operation.Types[typeRef].TypeKind == wgast.TypeKindNonNull {
					typeRef = operation.Types[typeRef].OfType
				}
				if operation.Types[typeRef].TypeKind == wgast.TypeKindList {
					kind := reflect.ValueOf(value).Kind()
					if kind != reflect.Slice && kind != reflect.Array {
						coerced[name] = []any{value}
					}
				}
			}
		}
		var err error
		variables, err = json.Marshal(coerced)
		if err != nil {
			return err
		}
	}
	return variablesvalidation.NewVariablesValidator(variablesvalidation.VariablesValidatorOptions{
		DisableExposingVariablesContent: true,
	}).Validate(operation, definition, variables)
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

	err = schemaManager.Generate(ctx, collections)
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
			return defrap.NewFilterFromString(gotSchemaManager.Definition(), collectionType, body)
		}
	}
	return defrap.NewFilterFromString(p.schemaManager.Definition(), collectionType, body)
}
