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
	"context"
	"regexp"
	"strings"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/tests/action"
)

var log = corelog.NewLogger("tests.multiplier")

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
					log.InfoContext(context.Background(),
						"Modified schema for secondary-index multiplier:\n"+newSchema)
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
var scalarTypes = []string{"String", "Int", "Float", "Float32", "Float64", "Boolean", "DateTime", "ID", "JSON"}

// scalarPatterns are precompiled patterns for scalar types.
var scalarPatterns = make([]*regexp.Regexp, len(scalarTypes))

// typeNamePattern extracts type names from "type TypeName { ... }" declarations.
var typeNamePattern = regexp.MustCompile(`type\s+(\w+)\s*\{`)

func init() {
	for i, typ := range scalarTypes {
		// Match scalar and array types:
		// - Type, Type!, [Type], [Type!], [Type]!, [Type!]!
		// The pattern handles all valid GraphQL type variations.
		// Uses word boundary (\b) after type name to avoid partial matches (e.g., Float matching Float32).
		scalarPatterns[i] = regexp.MustCompile(
			`(\w+:\s*)(\[?` + typ + `\b!?\]?!?)([^\n]*)(\n|$)`,
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

// hasSingleRelationTo checks if a type has a single (non-array) relation field pointing to targetType.
func hasSingleRelationTo(schema, sourceType, targetType string) bool {
	typeBlockPattern := regexp.MustCompile(`type\s+` + sourceType + `\s*\{([^}]*)\}`)
	match := typeBlockPattern.FindStringSubmatch(schema)
	if len(match) < 2 {
		return false
	}
	typeBody := match[1]

	singleRelPattern := regexp.MustCompile(`\w+:\s*` + targetType + `[^!\w\[]`)
	return singleRelPattern.MatchString(typeBody)
}

// isOneToOneRelation checks if there's a one-to-one relationship between two types.
// One-to-one exists when both types have single (non-array) relations to each other.
func isOneToOneRelation(schema, typeA, typeB string) bool {
	return hasSingleRelationTo(schema, typeA, typeB) && hasSingleRelationTo(schema, typeB, typeA)
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

	// Add @index to relation fields that hold foreign keys.
	// One-to-one relations are NOT indexed because DefraDB automatically creates a unique index.
	typeNames := extractTypeNames(schema)
	for _, typeName := range typeNames {
		result = addRelationIndexesForType(result, schema, typeName, typeNames)
	}

	return result
}

// addRelationIndexesForType adds @index to relation fields in the given type that hold foreign keys.
func addRelationIndexesForType(result, originalSchema, typeName string, allTypes []string) string {
	typeBlockPattern := regexp.MustCompile(`type\s+` + typeName + `\s*\{([^}]*)\}`)

	return typeBlockPattern.ReplaceAllStringFunc(result, func(typeBlock string) string {
		for _, otherType := range allTypes {
			// Skip one-to-one relations (DefraDB auto-creates unique index)
			if otherType != typeName && isOneToOneRelation(originalSchema, typeName, otherType) {
				continue
			}

			// Match single relation fields to otherType (not arrays)
			// Pattern: fieldName: OtherType (followed by space, @, newline, or end)
			pattern := regexp.MustCompile(`(\w+:\s*)(` + otherType + `)(\s|@|\n|$)`)
			typeBlock = pattern.ReplaceAllStringFunc(typeBlock, func(match string) string {
				if strings.Contains(match, "@index") {
					return match
				}
				return pattern.ReplaceAllString(match, "${1}${2} @index${3}")
			})
		}
		return typeBlock
	})
}
