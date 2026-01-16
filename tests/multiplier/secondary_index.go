// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package multiplier

import (
	"regexp"
	"strings"

	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/tests/action"
)

func init() {
	multiplier.Register(&secondaryIndex{})
}

// SecondaryIndex multiplier automatically adds @index directives to test schemas
// that don't already use indexes.
//
// This ensures query results are consistent regardless of whether indexes are present,
// by running existing tests with indexes enabled on all indexable fields.
const SecondaryIndex Name = "secondary-index"

type secondaryIndex struct{}

var _ Multiplier = (*secondaryIndex)(nil)
var _ multiplier.ActionAwareSkipper = (*secondaryIndex)(nil)

func (m *secondaryIndex) Name() Name {
	return SecondaryIndex
}

// ShouldSkip implements [multiplier.ActionAwareSkipper].
//
// Returns true if the action set contains index-related actions or schemas with
// existing @index directives, as these tests are specifically designed to test
// indexing behavior and should not be modified by this multiplier.
func (m *secondaryIndex) ShouldSkip(actions action.Actions) bool {
	if hasIndexActions(actions) {
		return true
	}

	for _, a := range actions {
		if schemaAdd, ok := a.(*action.AddSchema); ok {
			if hasIndexDirective(schemaAdd.Schema) {
				return true
			}
		}
	}

	return false
}

func (m *secondaryIndex) Apply(source action.Actions) action.Actions {
	result := make(action.Actions, len(source))
	modified := false

	for i, a := range source {
		if schemaAdd, ok := a.(*action.AddSchema); ok {
			if !hasIndexDirective(schemaAdd.Schema) {
				newSchema := addIndexesToSchema(schemaAdd.Schema)
				if newSchema != schemaAdd.Schema {
					newSchemaAdd := *schemaAdd
					newSchemaAdd.Schema = newSchema
					result[i] = &newSchemaAdd
					modified = true
					continue
				}
			}
		}
		result[i] = a
	}

	if !modified {
		return source
	}
	return result
}

// hasIndexActions returns true if any action in the set is index-related.
func hasIndexActions(actions action.Actions) bool {
	for _, a := range actions {
		switch a.(type) {
		case *action.CreateIndex, *action.DropIndex, *action.GetIndexes:
			return true
		}
	}
	return false
}

// hasIndexDirective returns true if the schema contains @index directive.
func hasIndexDirective(schema string) bool {
	return strings.Contains(schema, "@index")
}

// scalarTypes are the built-in types that can be indexed.
var scalarTypes = []string{"String", "Int", "Float", "Boolean", "DateTime", "ID", "JSON"}

// scalarPatterns are precompiled patterns for scalar types.
var scalarPatterns = make([]*regexp.Regexp, len(scalarTypes))

// typeNamePattern extracts type names from "type TypeName { ... }" declarations.
var typeNamePattern = regexp.MustCompile(`type\s+(\w+)\s*\{`)

func init() {
	for i, typ := range scalarTypes {
		// Match scalar and array types:
		// - Type, Type!, [Type], [Type!], [Type]!, [Type!]!
		// The pattern handles all valid GraphQL type variations
		scalarPatterns[i] = regexp.MustCompile(
			`(\w+:\s*)(\[?` + typ + `!?\]?!?)([^\n]*)(\n|$)`,
		)
	}
}

// extractTypeNames returns all type names defined in the schema.
func extractTypeNames(schema string) []string {
	matches := typeNamePattern.FindAllStringSubmatch(schema, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}
	return names
}

// hasArrayFieldOfType checks if the schema has an array field of the given type.
// For example, hasArrayFieldOfType(schema, "Device") returns true if schema contains "[Device]".
func hasArrayFieldOfType(schema, typeName string) bool {
	pattern := regexp.MustCompile(`:\s*\[` + typeName + `[!\]]`)
	return pattern.MatchString(schema)
}

// addIndexesToSchema adds @index directives to indexable fields (scalars, arrays, and relations).
// This function assumes the schema has no existing @index directives (checked by ShouldSkip/Apply).
func addIndexesToSchema(schema string) string {
	result := schema

	// Add @index to scalar types
	for i := range scalarTypes {
		pattern := scalarPatterns[i]
		// Add @index after the type (before any other directives)
		// Example: "name: String @crdt(...)\n" -> "name: String @index @crdt(...)\n"
		result = pattern.ReplaceAllString(result, "${1}${2} @index${3}${4}")
	}

	// Add @index to relation fields that hold foreign keys:
	typeNames := extractTypeNames(schema)
	for _, typeName := range typeNames {
		// Check if this type is the "many" side of a one-to-many relation
		// (i.e., there's an array field [ThisType] somewhere in the schema)
		if hasArrayFieldOfType(schema, typeName) {
			for _, otherType := range typeNames {
				if otherType == typeName {
					continue
				}
				// Match: fieldName: OtherType (single relation to a type that has [ThisType] array)
				// Array fields like [OtherType] won't match because the pattern requires
				// the type name immediately after the colon.
				pattern := regexp.MustCompile(`(\w+:\s*)(` + otherType + `)([^\n]*)(\n|$)`)
				result = pattern.ReplaceAllString(result, "${1}${2} @index${3}${4}")
			}
		}

		// One-to-one: the side with @primary holds the foreign key
		pattern := regexp.MustCompile(`(\w+:\s*)(` + typeName + `)([^\n]*@primary[^\n]*)(\n|$)`)
		result = pattern.ReplaceAllString(result, "${1}${2} @index${3}${4}")
	}

	return result
}
