// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// SchemaDefinition represents a schema definition.
type SchemaDefinition struct {
	// The name of this schema definition.
	Name string

	// The serialized definition of this schema definition.
	// todo: this needs to be properly typed and handled according to
	// https://github.com/sourcenetwork/defradb/issues/863
	Body []byte
}

// Parser represents the object responsible for handling stuff specific to a query language.
// This includes schema and query parsing, and introspection.
type Parser interface {
	// Returns true if the given string is an introspection request.
	IsIntrospection(request string) bool

	// Executes the given introspection request.
	ExecuteIntrospection(request string) *client.QueryResult

	// Parses the given request, returning a strongly typed model of that request.
	Parse(request string) (*request.Request, []error)

	// NewFilterFromString creates a new filter from a string.
	NewFilterFromString(collectionType string, body string) (immutable.Option[request.Filter], error)

	// Adds the given schema to this parser's model.
	AddSchema(ctx context.Context, schema string) error
}
