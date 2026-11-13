// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/graphql-go/language/kinds"
	"github.com/sourcenetwork/graphql-go/language/visitor"
)

// ValidateUniqueInputFieldNames checks that every input object literal in the
// document declares each of its fields at most once.
//
// The GraphQL specification requires the fields of an input object to be
// unique. The upstream validation rule stops descending as soon as it visits a
// field, so it only catches duplicates at the top level of an input object and
// silently accepts duplicates nested inside another field's value, for example
// `filter: {_or: {fieldName: {_eq: "a"}, fieldName: {_eq: "b"}}}`. When that
// happens the later value overwrites the earlier one and the query runs against
// a condition the user did not write.
//
// This walk visits every input object literal in the document, including nested
// ones, and reports each field that appears more than once.
func ValidateUniqueInputFieldNames(doc *ast.Document) []error {
	var errs []error
	visitor.Visit(doc, &visitor.VisitorOptions{
		KindFuncMap: map[string]visitor.NamedVisitFuncs{
			kinds.ObjectValue: {
				Kind: func(p visitor.VisitFuncParams) (string, any) {
					obj, ok := p.Node.(*ast.ObjectValue)
					if !ok {
						return visitor.ActionNoChange, nil
					}
					seen := make(map[string]struct{}, len(obj.Fields))
					for _, field := range obj.Fields {
						if field.Name == nil {
							continue
						}
						name := field.Name.Value
						if _, isDuplicate := seen[name]; isDuplicate {
							errs = append(errs, NewErrDuplicateInputField(name))
							continue
						}
						seen[name] = struct{}{}
					}
					return visitor.ActionNoChange, nil
				},
			},
		},
	}, nil)
	return errs
}
