// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	"fmt"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	schemaTypes "github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
)

// ParseRequest parses a root ast.Document, and returns a formatted Request object.
// Requires a non-nil doc, will error otherwise.
func ParseRequest(schema gql.Schema, doc *ast.Document, options *client.GQLOptions) (*request.Request, []error) {
	if doc == nil {
		return nil, []error{client.NewErrUninitializeProperty("ParseRequest", "doc")}
	}
	exe, err := gql.BuildExecutionContext(gql.BuildExecutionCtxParams{
		Schema:        schema,
		AST:           doc,
		OperationName: options.OperationName,
		Args:          options.Variables,
	})
	if err != nil {
		return nil, []error{err}
	}
	operationType, err := gql.GetOperationRootType(exe.Schema, exe.Operation)
	if err != nil {
		return nil, []error{err}
	}
	collectedFields := gql.CollectFields(gql.CollectFieldsParams{
		ExeContext:   exe,
		RuntimeType:  operationType,
		SelectionSet: exe.Operation.GetSelectionSet(),
	})

	r := &request.Request{
		Queries:      make([]*request.OperationDefinition, 0),
		Mutations:    make([]*request.OperationDefinition, 0),
		Subscription: make([]*request.OperationDefinition, 0),
	}

	astOpDef := exe.Operation.(*ast.OperationDefinition)
	switch exe.Operation.GetOperation() {
	case ast.OperationTypeQuery:
		parsedQueryOpDef, errs := parseQueryOperationDefinition(exe, collectedFields)
		if errs != nil {
			return nil, errs
		}
		parsedDirectives, err := parseDirectives(astOpDef.Directives)
		if err != nil {
			return nil, []error{err}
		}
		parsedQueryOpDef.Directives = parsedDirectives

		r.Queries = append(r.Queries, parsedQueryOpDef)

	case ast.OperationTypeMutation:
		parsedMutationOpDef, err := parseMutationOperationDefinition(exe, collectedFields)
		if err != nil {
			return nil, []error{err}
		}

		parsedDirectives, err := parseDirectives(astOpDef.Directives)
		if err != nil {
			return nil, []error{err}
		}
		parsedMutationOpDef.Directives = parsedDirectives

		r.Mutations = append(r.Mutations, parsedMutationOpDef)

	case ast.OperationTypeSubscription:
		parsedSubscriptionOpDef, err := parseSubscriptionOperationDefinition(exe, collectedFields)
		if err != nil {
			return nil, []error{err}
		}

		parsedDirectives, err := parseDirectives(astOpDef.Directives)
		if err != nil {
			return nil, []error{err}
		}
		parsedSubscriptionOpDef.Directives = parsedDirectives

		r.Subscription = append(r.Subscription, parsedSubscriptionOpDef)

	default:
		return nil, []error{ErrUnknownGQLOperation}
	}

	return r, nil
}

// parseDirectives returns all directives that were found if parsing and validation succeeds,
// otherwise returns the first error that is encountered.
func parseDirectives(astDirectives []*ast.Directive) (request.Directives, error) {
	// Set the default states of the directives if they aren't found and no error(s) occur.
	explainDirective := immutable.None[request.ExplainType]()

	// Iterate through all directives and ensure that the directive we find are validated.
	// - Note: the location we don't need to worry about as the schema takes care of it, as when
	//         request is made there will be a syntax error for directive usage at the wrong location,
	//         unless we add another directive with the same name, for example `@explain` is added
	//         at another location (which we must avoid).
	for _, astDirective := range astDirectives {
		if astDirective == nil {
			return request.Directives{}, errors.New("found a nil directive in the AST")
		}

		if astDirective.Name.Value == request.ExplainLabel {
			// Explain directive found, lets parse and validate the directive.
			parsedExplainDirective, err := parseExplainDirective(astDirective)
			if err != nil {
				return request.Directives{}, err
			}
			explainDirective = parsedExplainDirective
		}
	}

	return request.Directives{
		ExplainType: explainDirective,
	}, nil
}

// parseExplainDirective parses the explain directive AST and returns an error if the parsing or
// validation goes wrong, otherwise returns the parsed explain type information.
func parseExplainDirective(astDirective *ast.Directive) (immutable.Option[request.ExplainType], error) {
	if len(astDirective.Arguments) == 0 {
		return immutable.Some(request.SimpleExplain), nil
	}

	if len(astDirective.Arguments) != 1 {
		return immutable.None[request.ExplainType](), ErrInvalidNumberOfExplainArgs
	}

	arg := astDirective.Arguments[0]
	if arg.Name.Value != schemaTypes.ExplainArgNameType {
		return immutable.None[request.ExplainType](), ErrInvalidExplainTypeArg
	}

	switch arg.Value.GetValue() {
	case schemaTypes.ExplainArgSimple:
		return immutable.Some(request.SimpleExplain), nil

	case schemaTypes.ExplainArgExecute:
		return immutable.Some(request.ExecuteExplain), nil

	case schemaTypes.ExplainArgDebug:
		return immutable.Some(request.DebugExplain), nil

	default:
		return immutable.None[request.ExplainType](), ErrUnknownExplainType
	}
}

func getFieldAlias(field *ast.Field) immutable.Option[string] {
	if field.Alias == nil {
		return immutable.None[string]()
	}
	return immutable.Some(field.Alias.Value)
}

func parseSelectFields(
	exe *gql.ExecutionContext,
	parent *gql.Object,
	fields *ast.SelectionSet,
) ([]request.Selection, error) {
	var selections []request.Selection
	for _, selection := range fields.Selections {
		switch node := selection.(type) {
		case *ast.FragmentSpread:
			fragment, ok := exe.Fragments[node.Name.Value]
			if !ok {
				return nil, fmt.Errorf("fragment not found %s", node.Name.Value)
			}
			selection, err := parseSelectFields(exe, parent, fragment.GetSelectionSet())
			if err != nil {
				return nil, err
			}
			selections = append(selections, selection...)

		case *ast.Field:
			var selection request.Selection
			if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				s, err := parseAggregate(exe, parent, node)
				if err != nil {
					return nil, err
				}
				selection = s
			} else if node.SelectionSet == nil { // regular field
				selection = parseField(node)
			} else { // sub type with extra fields
				s, err := parseSelect(exe, parent, node)
				if err != nil {
					return nil, err
				}
				selection = s
			}
			selections = append(selections, selection)
		}
	}

	return selections, nil
}

// parseField simply parses the Name/Alias
// into a Field type
func parseField(field *ast.Field) *request.Field {
	return &request.Field{
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}
}

func getArgumentType(field *gql.FieldDefinition, name string) (gql.Input, bool) {
	for _, arg := range field.Args {
		if arg.Name() == name {
			return arg.Type, true
		}
	}
	return nil, false
}

// typeFromFieldDef will return the output gql.Object type from the given field.
// The return type may be a gql.Object or a gql.List, if it is a List type, we
// need to get the concrete "OfType".
func typeFromFieldDef(field *gql.FieldDefinition) (*gql.Object, error) {
	var fieldObject *gql.Object
	switch ftype := field.Type.(type) {
	case *gql.Object:
		fieldObject = ftype
	case *gql.List:
		fieldObject = ftype.OfType.(*gql.Object)
	default:
		return nil, client.NewErrUnhandledType("field", field)
	}
	return fieldObject, nil
}
