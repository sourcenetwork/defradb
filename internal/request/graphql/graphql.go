// Copyright 2024 Democratized Data Foundation
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
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

var (
	//go:embed base.graphql
	baseGraphql string
	//go:embed commit.graphql
	commitGraphql string
	//go:embed schema.graphql
	schemaTemplateGraphql string
	// schemaTemplate is the template for generating new schema types
	schemaTemplate = template.Must(template.New("").Funcs(schemaTemplateFuncs).Parse(schemaTemplateGraphql))
)

// schemaTemplateFuncs is a map of functions available in the schema template.
var schemaTemplateFuncs = template.FuncMap{
	"IsViewObject": func(collection client.CollectionDefinition) bool {
		return !collection.Description.Name.HasValue() || len(collection.Description.QuerySources()) > 0
	},
	"IsSystemField": func(field client.FieldDefinition) bool {
		return strings.HasPrefix(field.Name, "_")
	},
	"IsDocIDField": func(field client.FieldDefinition) bool {
		return field.Name == request.DocIDFieldName
	},
	"IsNumericField": func(field client.FieldDefinition) bool {
		var kind string
		if field.Kind.IsArray() {
			kind = field.Kind.Underlying()
		} else {
			kind = field.Kind.String()
		}
		return kind == "Int" || kind == "Float"
	},
	"FieldOperatorBlock": func(field client.FieldDefinition) string {
		return kindToOperatorBlock[field.Kind]
	},
}

// kindToOperatorBlock is a mapping of FieldKinds to input types used in filter blocks.
var kindToOperatorBlock = map[client.FieldKind]string{
	client.FieldKind_DocID:                 "IDOperatorBlock",
	client.FieldKind_BOOL_ARRAY:            "NotNullBooleanOperatorBlock",
	client.FieldKind_FLOAT_ARRAY:           "NotNullFloatOperatorBlock",
	client.FieldKind_INT_ARRAY:             "NotNullIntOperatorBlock",
	client.FieldKind_STRING_ARRAY:          "NotNullStringOperatorBlock",
	client.FieldKind_NILLABLE_BLOB:         "BlobOperatorBlock",
	client.FieldKind_NILLABLE_BOOL:         "BooleanOperatorBlock",
	client.FieldKind_NILLABLE_BOOL_ARRAY:   "BooleanOperatorBlock",
	client.FieldKind_NILLABLE_DATETIME:     "DateTimeOperatorBlock",
	client.FieldKind_NILLABLE_FLOAT:        "FloatOperatorBlock",
	client.FieldKind_NILLABLE_FLOAT_ARRAY:  "FloatOperatorBlock",
	client.FieldKind_NILLABLE_INT:          "IntOperatorBlock",
	client.FieldKind_NILLABLE_INT_ARRAY:    "IntOperatorBlock",
	client.FieldKind_NILLABLE_JSON:         "JSONOperatorBlock",
	client.FieldKind_NILLABLE_STRING:       "StringOperatorBlock",
	client.FieldKind_NILLABLE_STRING_ARRAY: "StringOperatorBlock",
}

// GenerateSchema returns a GraphQL schema containing all defined types
// including the user generated types defined in the given collections.
func GenerateSchema(cols []client.CollectionDefinition) (string, error) {
	data := make(map[string]client.CollectionDefinition)
	for _, col := range cols {
		if col.Description.Name.HasValue() {
			data[col.Description.Name.Value()] = col
		} else {
			data[col.Schema.Name] = col
		}
	}
	var out bytes.Buffer
	if err := schemaTemplate.Execute(&out, data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%s", baseGraphql, commitGraphql, out.String()), nil
}
