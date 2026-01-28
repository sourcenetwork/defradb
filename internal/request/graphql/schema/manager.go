// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	gql "github.com/sourcenetwork/graphql-go"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	schema    gql.Schema
	Generator *Generator
}

// NewSchemaManager returns a new instance of a SchemaManager
// with a new default type map
func NewSchemaManager(isSearchableEncryptionEnabled bool) (*SchemaManager, error) {
	schema, err := defaultSchema()
	if err != nil {
		return nil, err
	}
	sm := &SchemaManager{
		schema: schema,
	}
	sm.NewGenerator(isSearchableEncryptionEnabled)
	return sm, nil
}

func (s *SchemaManager) Schema() *gql.Schema {
	return &s.schema
}

// ResolveTypes ensures all necessary types are defined, and
// resolves any remaining thunks/closures defined on object fields.
// It should be called *after* all dependent types have been added.
func (s *SchemaManager) ResolveTypes() error {
	// basically, this function just refreshes the
	// schema.TypeMap, and runs the internal
	// typeMapReducer (https://github.com/sourcenetwork/graphql-go/blob/v0.7.9/schema.go#L275)
	// which ensures all the necessary types are defined in the
	// typeMap, and if there are any outstanding Thunks/closures
	// resolve them.

	// ATM, there is no function to easily call the internal
	// typeMapReducer function, so as a hack, we are just
	// going to re-add the Query type.

	for _, gqlType := range s.schema.TypeMap() {
		object, isObject := gqlType.(*gql.Object)
		if !isObject {
			continue
		}
		// We need to make sure the object's fields are resolved
		object.Fields()

		if object.Error() != nil {
			return object.Error()
		}
	}

	query := s.schema.QueryType()
	return s.schema.AppendType(query)
}

func (s *SchemaManager) ParseSDL(sdl string) ([]core.Collection, error) {
	src := source.NewSource(&source.Source{
		Body: []byte(sdl),
	})
	doc, err := gqlp.Parse(gqlp.ParseParams{
		Source: src,
	})
	if err != nil {
		return nil, err
	}
	// The user provided SDL must be validated using the latest generated schema
	// so that relations to other user defined types do not return an error.
	validation := gql.ValidateDocument(&s.schema, doc, gql.SpecifiedRules)
	if !validation.IsValid {
		for _, e := range validation.Errors {
			err = errors.Join(err, e)
		}
		return nil, err
	}
	return fromAst(doc)
}

func (s *SchemaManager) WriteSDL(writer io.Writer) error {
	params := gql.Params{Schema: *s.Schema(), RequestString: introspectionQueryRequest}
	r := gql.Do(params)
	if len(r.Errors) != 0 {
		// for simplicity we're just going to return the
		// first error, if there are more, they'll be caught on
		// follow up invocations.
		return errors.Join(ErrGeneratingSDL, r.Errors[0])
	}

	respJson, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	respBuf := bytes.NewBuffer(respJson)

	converter := introspection.JsonConverter{}
	doc, err := converter.GraphQLDocument(respBuf)
	if err != nil {
		return err
	}

	err = astprinter.PrintIndent(doc, []byte("    "), writer)
	if err != nil {
		return errors.Join(ErrWritingSDL, err)
	}
	return nil
}

const introspectionQueryRequest = "query IntrospectionQuery{__schema{queryType{name}mutationType{name}subscriptionType{name}types{...FullType}directives{name description locations args{...InputValue}}}}fragment FullType on __Type{kind name description fields(includeDeprecated:true){name description args{...InputValue}type{...TypeRef}isDeprecated deprecationReason}inputFields{...InputValue}interfaces{...TypeRef}enumValues(includeDeprecated:true){name description isDeprecated deprecationReason}possibleTypes{...TypeRef}}fragment InputValue on __InputValue{name description type{...TypeRef}defaultValue}fragment TypeRef on __Type{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name}}}}}}}}}}" //nolint:lll
