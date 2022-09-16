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
	"context"
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/source"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"

	gql "github.com/graphql-go/graphql"
	gqlp "github.com/graphql-go/graphql/language/parser"

	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
	schemaTypes "github.com/sourcenetwork/defradb/query/graphql/schema/types"
)

// Given a basic developer defined schema in GraphQL Schema Definition Language
// create a fully DefraDB complaint GraphQL schema using a "code-first" dynamic
// approach

// Type represents a developer defined type, and its associated graphQL generated types
type Type struct {
	gql.ObjectConfig
	Object *gql.Object
}

// Generator creates all the necessary typed schema definitions from an AST Document
// and adds them to the Schema via the SchemaManager
type Generator struct {
	typeDefs []*gql.Object
	manager  *SchemaManager

	expandedFields map[string]bool
}

// NewGenerator creates a new instance of the Generator
// from a given SchemaManager
func (m *SchemaManager) NewGenerator() *Generator {
	m.Generator = &Generator{
		manager:        m,
		expandedFields: make(map[string]bool),
	}
	return m.Generator
}

// FromSDL generates the query type definitions from a
// encoded GraphQL Schema Definition Language string
func (g *Generator) FromSDL(
	ctx context.Context,
	schema string,
) ([]*gql.Object, *ast.Document, error) {
	// parse to AST
	source := source.NewSource(&source.Source{
		Body: []byte(schema),
	})
	doc, err := gqlp.Parse(gqlp.ParseParams{
		Source: source,
	})
	if err != nil {
		return nil, nil, err
	}
	// generate from AST
	types, err := g.FromAST(ctx, doc)
	return types, doc, err
}

func (g *Generator) FromAST(ctx context.Context, document *ast.Document) ([]*gql.Object, error) {
	typeMapBeforeMutation := g.manager.schema.TypeMap()
	typesBeforeMutation := make(map[string]any, len(typeMapBeforeMutation))

	for typeName := range typeMapBeforeMutation {
		typesBeforeMutation[typeName] = struct{}{}
	}

	result, err := g.fromAST(ctx, document)

	if err != nil {
		// - If there is an error we should drop any new objects as they may be partial, polluting
		//   the in-memory cache.
		// - This is quite a simple check at the moment (on type name) - this should be expanded
		//   when we allow schema mutation/deletion.
		// - There is no guarantee that `typeMapBeforeMutation` will still be the object returned
		//   by `schema.TypeMap()`, so we should re-fetch it
		typeMapAfterMutation := g.manager.schema.TypeMap()
		for typeName := range typeMapAfterMutation {
			if _, typeExistedBeforeMutation := typesBeforeMutation[typeName]; !typeExistedBeforeMutation {
				delete(typeMapAfterMutation, typeName)
			}
		}

		return nil, err
	}

	return result, nil
}

// FromAST generates the query type definitions from a
// parsed GraphQL Schema Definition Language AST document
func (g *Generator) fromAST(ctx context.Context, document *ast.Document) ([]*gql.Object, error) {
	// build base types
	defs, err := g.buildTypesFromAST(ctx, document)
	if err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// for each built type generate query inputs
	queryType := g.manager.schema.QueryType()
	generatedQueryFields := make([]*gql.Field, 0)
	for _, t := range g.typeDefs {
		f, err := g.GenerateQueryInputForGQLType(ctx, t)
		if err != nil {
			return nil, err
		}
		queryType.AddFieldConfig(f.Name, f)
		generatedQueryFields = append(generatedQueryFields, f)
	}

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	if err := g.genAggregateFields(ctx); err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	generatedFilterLeafArgs := []*gql.InputObject{}
	for _, defaultType := range inlineArrayTypes() {
		leafFilterArg := g.genLeafFilterArgInput(defaultType)
		generatedFilterLeafArgs = append(generatedFilterLeafArgs, leafFilterArg)
	}

	for _, t := range generatedFilterLeafArgs {
		err := g.appendIfNotExists(t)
		if err != nil {
			return nil, err
		}
	}

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// secondary pass to expand query collection type
	// argument inputs
	// query := g.manager.schema.QueryType()
	// queries := query.Fields()
	// only apply to generated query fields, and only once
	for _, def := range generatedQueryFields {
		switch obj := def.Type.(type) {
		case *gql.List:
			if err := g.expandInputArgument(obj.OfType.(*gql.Object)); err != nil {
				return nil, err
			}
		case *gql.Scalar:
			if _, isAggregate := parserTypes.Aggregates[def.Name]; isAggregate {
				for name, aggregateTarget := range def.Args {
					expandedField := &gql.InputObjectFieldConfig{
						Type: g.manager.schema.TypeMap()[name+"FilterArg"],
					}
					aggregateTarget.Type.(*gql.InputObject).AddFieldConfig(parserTypes.FilterClause, expandedField)
				}
			}
		}
	}

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// now let's generate the mutation types.
	mutationType := g.manager.schema.MutationType()
	for _, t := range g.typeDefs {
		fs, err := g.GenerateMutationInputForGQLType(t)
		if err != nil {
			return nil, err
		}
		for _, f := range fs { // GenMutation returns multiple fields to be added
			mutationType.AddFieldConfig(f.Name, f)
		}
	}

	// final resolve
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	return defs, nil
}

func (g *Generator) expandInputArgument(obj *gql.Object) error {
	fields := obj.Fields()
	for f, def := range fields {
		// ignore reserved fields, execpt the Group field (as that requires typing), and aggregates
		if _, ok := parserTypes.ReservedFields[f]; ok && f != parserTypes.GroupFieldName {
			if _, isAggregate := parserTypes.Aggregates[f]; !isAggregate {
				continue
			}
		}
		// Both the object name and the field name should be used as the key
		// in case the child object type is referenced multiple times from the same parent type
		fieldKey := obj.Name() + f
		switch t := def.Type.(type) {
		case *gql.Object:
			if _, complete := g.expandedFields[fieldKey]; complete {
				continue
			}
			g.expandedFields[fieldKey] = true

			// make sure all the sub fields are expanded first
			if err := g.expandInputArgument(t); err != nil {
				return err
			}

			// new field object with arguments (single)
			expandedField, err := g.createExpandedFieldSingle(def, t)
			if err != nil {
				return err
			}

			obj.AddFieldConfig(f, expandedField)

		case *gql.List: // new field object with arguments (list)
			listType := t.OfType
			if _, complete := g.expandedFields[fieldKey]; complete {
				continue
			}
			g.expandedFields[fieldKey] = true

			if listObjType, ok := listType.(*gql.Object); ok {
				if err := g.expandInputArgument(listObjType); err != nil {
					return err
				}

				expandedField, err := g.createExpandedFieldList(def, listObjType)
				if err != nil {
					return err
				}
				obj.AddFieldConfig(f, expandedField)
			}
		case *gql.Scalar:
			if _, isAggregate := parserTypes.Aggregates[f]; isAggregate {
				if err := g.createExpandedFieldAggregate(obj, def); err != nil {
					return err
				}
			}
			// @todo: check if NonNull is possible here
			//case *gql.NonNull:
			// get subtype
		}
	}

	return nil
}

func (g *Generator) createExpandedFieldAggregate(
	obj *gql.Object,
	f *gql.FieldDefinition,
) error {
	for _, aggregateTarget := range f.Args {
		target := aggregateTarget.Name()
		var filterTypeName string
		if target == parserTypes.GroupFieldName {
			filterTypeName = obj.Name() + "FilterArg"
		} else {
			if targeted := obj.Fields()[target]; targeted != nil {
				if list, isList := targeted.Type.(*gql.List); isList && gql.IsLeafType(list.OfType) {
					// If it is a list of leaf types - the filter is just the set of OperatorBlocks
					// that are supported by this type - there can be no field selections.
					if notNull, isNotNull := list.OfType.(*gql.NonNull); isNotNull {
						// GQL does not support '!' in type names, and so we have to manipulate the
						// underlying name like this if it is a nullable type.
						filterTypeName = fmt.Sprintf("NotNull%sFilterArg", notNull.OfType.Name())
					} else {
						filterTypeName = genTypeName(list.OfType, "FilterArg")
					}
				} else {
					filterTypeName = targeted.Type.Name() + "FilterArg"
				}
			} else {
				return errors.New(fmt.Sprintf(
					"Aggregate target not found. HostObject: {%s}, Target: {%s}",
					obj.Name(),
					target,
				))
			}
		}

		if filterType, canHaveFilter := g.manager.schema.TypeMap()[filterTypeName]; canHaveFilter {
			// Sometimes a filter is not permitted, for example when aggregating `_version`
			expandedField := &gql.InputObjectFieldConfig{
				Type: filterType,
			}
			aggregateTarget.Type.(*gql.InputObject).AddFieldConfig("filter", expandedField)
		}
	}

	return nil
}

func (g *Generator) createExpandedFieldSingle(
	f *gql.FieldDefinition,
	t *gql.Object,
) (*gql.Field, error) {
	typeName := t.Name()
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: f.Name,
		Type: t,
		Args: gql.FieldConfigArgument{
			"filter": schemaTypes.NewArgConfig(g.manager.schema.TypeMap()[typeName+"FilterArg"]),
		},
	}
	return field, nil
}

// @todo: add field reference so we can copy extra fields (like description, depreciation, etc)
func (g *Generator) createExpandedFieldList(
	f *gql.FieldDefinition,
	t *gql.Object,
) (*gql.Field, error) {
	typeName := t.Name()
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: f.Name,
		Type: gql.NewList(t),
		Args: gql.FieldConfigArgument{
			"filter": schemaTypes.NewArgConfig(g.manager.schema.TypeMap()[typeName+"FilterArg"]),
			"groupBy": schemaTypes.NewArgConfig(
				gql.NewList(gql.NewNonNull(g.manager.schema.TypeMap()[typeName+"Fields"])),
			),
			"order":                  schemaTypes.NewArgConfig(g.manager.schema.TypeMap()[typeName+"OrderArg"]),
			parserTypes.LimitClause:  schemaTypes.NewArgConfig(gql.Int),
			parserTypes.OffsetClause: schemaTypes.NewArgConfig(gql.Int),
		},
	}

	return field, nil
}

// @todo: Add Schema Directives (IE: relation, etc..)

// @todo: Add validation support for the AST
// @body: Type generation is only supported for Object type definitions.
// Unions, Interfaces, etc are not currently supported.

// Given a parsed AST of  developer defined types
// extract and return the correct gql.Object type(s)
func (g *Generator) buildTypesFromAST(
	ctx context.Context,
	document *ast.Document,
) ([]*gql.Object, error) {
	// @todo: Check for duplicate named defined types in the TypeMap
	// get all the defined types from the AST
	objs := make([]*gql.Object, 0)

	for _, def := range document.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			// check if type exists
			if _, ok := g.manager.schema.TypeMap()[defType.Name.Value]; ok {
				return nil, errors.New(fmt.Sprintf("Schema type already exists: %s", defType.Name.Value))
			}

			objconf := gql.ObjectConfig{}
			// otype.astDef = defType // keep a reference
			if defType.Name != nil {
				objconf.Name = defType.Name.Value
			}
			if defType.Description != nil {
				objconf.Description = defType.Description.Value
			}

			// Wrap field definition in a thunk so we can
			// handle any embedded object which is defined
			// at a future point in time.
			fieldsThunk := (gql.FieldsThunk)(func() (gql.Fields, error) {
				fields := gql.Fields{}

				// @todo: Check if this is a collection (relation) type
				// or just a embedded only type (which doesn't need a key)
				// automatically add the _key: ID field to the type
				fields[parserTypes.DocKeyFieldName] = &gql.Field{Type: gql.ID}

				for _, field := range defType.Fields {
					fType := new(gql.Field)
					if field.Name != nil {
						fType.Name = field.Name.Value
					}
					if field.Description != nil {
						fType.Description = field.Description.Value
					}

					t := field.Type
					ttype, err := astNodeToGqlType(g.manager.schema.TypeMap(), t)
					if err != nil {
						return nil, err
					}

					// check if ttype is a Object value
					// if so, add appropriate relationship data
					// @todo check various directives for nature
					// of object relationship
					switch subobj := ttype.(type) {
					case *gql.Object:
						fields[fType.Name+"_id"] = &gql.Field{Type: gql.ID}

						// register the relation
						relName, err := getRelationshipName(field, objconf, ttype)
						if err != nil {
							return nil, err
						}

						// set the primary relation bit on the relation type if the directive exists on the
						// field
						relType := client.Relation_Type_ONE
						if _, exists := findDirective(field, "primary"); exists {
							relType |= client.Relation_Type_Primary
						}

						_, err = g.manager.Relations.RegisterSingle(
							relName,
							ttype.Name(),
							fType.Name,
							relType,
						)
						if err != nil {
							log.ErrorE(ctx, "Error while registering single relation for object", err)
						}

					case *gql.List:
						ltype := subobj.OfType
						if !gql.IsLeafType(ltype) {
							// We don't want to do this for inline-arrays (IsLeafType)
							relName, err := getRelationshipName(field, objconf, ltype)
							if err != nil {
								return nil, err
							}

							_, err = g.manager.Relations.RegisterSingle(
								relName,
								ltype.Name(),
								fType.Name,
								client.Relation_Type_MANY,
							)
							if err != nil {
								log.ErrorE(ctx, "Error while registering single relation for list", err)
							}
						}
					}

					fType.Type = ttype
					fields[fType.Name] = fType
				}

				// add _version field
				fields["_version"] = &gql.Field{
					Type: gql.NewList(schemaTypes.CommitObject),
				}

				gqlType, ok := g.manager.schema.TypeMap()[defType.Name.Value]
				if !ok {
					return nil, errors.New(fmt.Sprintf(
						"object not found whilst executing fields thunk: %s",
						defType.Name.Value,
					))
				}

				fields[parserTypes.GroupFieldName] = &gql.Field{
					Type: gql.NewList(gqlType),
				}

				return fields, nil
			})

			objconf.Fields = fieldsThunk

			obj := gql.NewObject(objconf)
			objs = append(objs, obj)
		}
	}

	// add all the new types now that they're converted to gql.Objects
	for _, obj := range objs {
		g.manager.schema.TypeMap()[obj.Name()] = obj
		g.typeDefs = append(g.typeDefs, obj)
	}

	return objs, nil
}

// Gets the name of the relationship. Will return the provided name if one is specified,
// otherwise will generate one
func getRelationshipName(
	field *ast.FieldDefinition,
	hostName gql.ObjectConfig,
	targetName gql.Type,
) (string, error) {
	// search for a @relation directive name, and return it if found
	for _, directive := range field.Directives {
		if directive.Name.Value == "relation" {
			for _, argument := range directive.Arguments {
				if argument.Name.Value == "name" {
					name, isString := argument.Value.GetValue().(string)
					if !isString {
						return "", errors.New(fmt.Sprintf(
							"Relationship name must be of type string, but was: %v",
							argument.Value.GetKind(),
						))
					}
					return name, nil
				}
			}
		}
	}

	// if no name is provided, generate one
	return genRelationName(hostName.Name, targetName.Name())
}

func (g *Generator) genAggregateFields(ctx context.Context) error {
	topLevelCountInputs := map[string]*gql.InputObject{}
	topLevelNumericAggInputs := map[string]*gql.InputObject{}

	for _, t := range g.typeDefs {
		numArg := g.genNumericAggregateBaseArgInputs(t)
		topLevelNumericAggInputs[t.Name()] = numArg
		// All base types need to be appended to the schema before calling genSumFieldConfig
		err := g.appendIfNotExists(numArg)
		if err != nil {
			return err
		}

		numericInlineArrayInputs := g.genNumericInlineArraySelectorObject(t)
		for _, obj := range numericInlineArrayInputs {
			err = g.appendIfNotExists(obj)
			if err != nil {
				return err
			}
		}

		obj := g.genCountBaseArgInputs(t)
		topLevelCountInputs[t.Name()] = obj
		err = g.appendIfNotExists(obj)
		if err != nil {
			return err
		}

		countableInlineArrayInputs := g.genCountInlineArrayInputs(t)
		for _, obj := range countableInlineArrayInputs {
			err = g.appendIfNotExists(obj)
			if err != nil {
				return err
			}
		}
	}

	for _, t := range g.typeDefs {
		countField, err := g.genCountFieldConfig(t)
		if err != nil {
			return err
		}
		t.AddFieldConfig(countField.Name, &countField)

		sumField, err := g.genSumFieldConfig(t)
		if err != nil {
			return err
		}
		t.AddFieldConfig(sumField.Name, &sumField)

		averageField, err := g.genAverageFieldConfig(t)
		if err != nil {
			return err
		}
		t.AddFieldConfig(averageField.Name, &averageField)
	}

	queryType := g.manager.schema.QueryType()

	topLevelCountField := genTopLevelCount(topLevelCountInputs)
	queryType.AddFieldConfig(topLevelCountField.Name, topLevelCountField)

	for _, topLevelAgg := range genTopLevelNumericAggregates(topLevelNumericAggInputs) {
		queryType.AddFieldConfig(topLevelAgg.Name, topLevelAgg)
	}

	return nil
}

func genTopLevelCount(topLevelCountInputs map[string]*gql.InputObject) *gql.Field {
	topLevelCountField := gql.Field{
		Name: parserTypes.CountFieldName,
		Type: gql.Int,
		Args: gql.FieldConfigArgument{},
	}

	for name, inputObject := range topLevelCountInputs {
		topLevelCountField.Args[name] = schemaTypes.NewArgConfig(inputObject)
	}

	return &topLevelCountField
}

func genTopLevelNumericAggregates(topLevelNumericAggInputs map[string]*gql.InputObject) []*gql.Field {
	topLevelSumField := gql.Field{
		Name: parserTypes.SumFieldName,
		Type: gql.Float,
		Args: gql.FieldConfigArgument{},
	}

	topLevelAverageField := gql.Field{
		Name: parserTypes.AverageFieldName,
		Type: gql.Float,
		Args: gql.FieldConfigArgument{},
	}

	for name, inputObject := range topLevelNumericAggInputs {
		topLevelSumField.Args[name] = schemaTypes.NewArgConfig(inputObject)
		topLevelAverageField.Args[name] = schemaTypes.NewArgConfig(inputObject)
	}

	return []*gql.Field{&topLevelSumField, &topLevelAverageField}
}

func (g *Generator) genCountFieldConfig(obj *gql.Object) (gql.Field, error) {
	childTypesByFieldName := map[string]gql.Type{}

	for _, field := range obj.Fields() {
		// Only lists can be counted
		if _, isList := field.Type.(*gql.List); !isList {
			continue
		}

		inputObjectName := genObjectCountName(field.Type.Name())
		countableObject, isSubTypeCountableCollection := g.manager.schema.TypeMap()[inputObjectName]
		if !isSubTypeCountableCollection {
			inputObjectName = genNumericInlineArrayCountName(obj.Name(), field.Name)
			var isSubTypeCountableInlineArray bool
			countableObject, isSubTypeCountableInlineArray = g.manager.schema.TypeMap()[inputObjectName]
			if !isSubTypeCountableInlineArray {
				continue
			}
		}

		childTypesByFieldName[field.Name] = countableObject
	}

	field := gql.Field{
		Name: parserTypes.CountFieldName,
		Type: gql.Int,
		Args: gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject)
	}

	return field, nil
}

func (g *Generator) genSumFieldConfig(obj *gql.Object) (gql.Field, error) {
	childTypesByFieldName := map[string]gql.Type{}

	for _, field := range obj.Fields() {
		// we can only sum list items
		listType, isList := field.Type.(*gql.List)
		if !isList {
			continue
		}

		var inputObjectName string
		if isNumericArray(listType) {
			inputObjectName = genNumericInlineArraySelectorName(obj.Name(), field.Name)
		} else {
			inputObjectName = genNumericObjectSelectorName(field.Type.Name())
		}

		subSumType, isSubTypeSumable := g.manager.schema.TypeMap()[inputObjectName]
		// If the item is not in the type map, it must contain no summable
		//  fields (e.g. no Int/Floats)
		if !isSubTypeSumable {
			continue
		}
		childTypesByFieldName[field.Name] = subSumType
	}

	field := gql.Field{
		Name: parserTypes.SumFieldName,
		Type: gql.Float,
		Args: gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject)
	}

	return field, nil
}

func (g *Generator) genAverageFieldConfig(obj *gql.Object) (gql.Field, error) {
	childTypesByFieldName := map[string]gql.Type{}

	for _, field := range obj.Fields() {
		// we can only sum list items
		listType, isList := field.Type.(*gql.List)
		if !isList {
			continue
		}

		var inputObjectName string
		if isNumericArray(listType) {
			inputObjectName = genNumericInlineArraySelectorName(obj.Name(), field.Name)
		} else {
			inputObjectName = genNumericObjectSelectorName(field.Type.Name())
		}

		subAverageType, isSubTypeAveragable := g.manager.schema.TypeMap()[inputObjectName]
		// If the item is not in the type map, it must contain no averagable
		//  fields (e.g. no Int/Floats)
		if !isSubTypeAveragable {
			continue
		}
		childTypesByFieldName[field.Name] = subAverageType
	}

	field := gql.Field{
		Name: parserTypes.AverageFieldName,
		Type: gql.Float,
		Args: gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject)
	}

	return field, nil
}

func (g *Generator) genNumericInlineArraySelectorObject(obj *gql.Object) []*gql.InputObject {
	objects := []*gql.InputObject{}
	for _, field := range obj.Fields() {
		// we can only act on list items
		listType, isList := field.Type.(*gql.List)
		if !isList {
			continue
		}

		if isNumericArray(listType) {
			// If it is an inline scalar array then we require an empty
			//  object as an argument due to the lack of union input types
			selectorObject := gql.NewInputObject(gql.InputObjectConfig{
				Name: genNumericInlineArraySelectorName(obj.Name(), field.Name),
				Fields: gql.InputObjectConfigFieldMap{
					parserTypes.LimitClause: &gql.InputObjectFieldConfig{
						Type:        gql.Int,
						Description: "The maximum number of child items to aggregate.",
					},
					parserTypes.OffsetClause: &gql.InputObjectFieldConfig{
						Type:        gql.Int,
						Description: "The index from which to start aggregating items.",
					},
					parserTypes.OrderClause: &gql.InputObjectFieldConfig{
						Type:        g.manager.schema.TypeMap()["Ordering"],
						Description: "The order in which to aggregate items.",
					},
				},
			})

			objects = append(objects, selectorObject)
		}
	}
	return objects
}

func genNumericObjectSelectorName(hostName string) string {
	return fmt.Sprintf("%s__%s", hostName, "NumericSelector")
}

func genNumericInlineArraySelectorName(hostName string, fieldName string) string {
	return fmt.Sprintf("%s__%s__%s", hostName, fieldName, "NumericSelector")
}

func (g *Generator) genCountBaseArgInputs(obj *gql.Object) *gql.InputObject {
	countableObject := gql.NewInputObject(gql.InputObjectConfig{
		Name: genObjectCountName(obj.Name()),
		Fields: gql.InputObjectConfigFieldMap{
			parserTypes.LimitClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: "The maximum number of child items to count.",
			},
			parserTypes.OffsetClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: "The index from which to start counting items.",
			},
		},
	})

	return countableObject
}

func (g *Generator) genCountInlineArrayInputs(obj *gql.Object) []*gql.InputObject {
	objects := []*gql.InputObject{}
	for _, field := range obj.Fields() {
		// we can only act on list items
		_, isList := field.Type.(*gql.List)
		if !isList {
			continue
		}

		// If it is an inline scalar array then we require an empty
		//  object as an argument due to the lack of union input types
		selectorObject := gql.NewInputObject(gql.InputObjectConfig{
			Name: genNumericInlineArrayCountName(obj.Name(), field.Name),
			Fields: gql.InputObjectConfigFieldMap{
				parserTypes.LimitClause: &gql.InputObjectFieldConfig{
					Type:        gql.Int,
					Description: "The maximum number of child items to count.",
				},
				parserTypes.OffsetClause: &gql.InputObjectFieldConfig{
					Type:        gql.Int,
					Description: "The index from which to start counting items.",
				},
			},
		})

		objects = append(objects, selectorObject)
	}
	return objects
}

func genNumericInlineArrayCountName(hostName string, fieldName string) string {
	return fmt.Sprintf("%s__%s__%s", hostName, fieldName, "CountSelector")
}

func genObjectCountName(hostName string) string {
	return fmt.Sprintf("%s__%s", hostName, "CountSelector")
}

// Generates the base (numeric-only) aggregate input object-type for the give gql object,
// declaring which fields are available for aggregation.
func (g *Generator) genNumericAggregateBaseArgInputs(obj *gql.Object) *gql.InputObject {
	var fieldThunk gql.InputObjectConfigFieldMapThunk = func() (gql.InputObjectConfigFieldMap, error) {
		fieldsEnum, enumExists := g.manager.schema.TypeMap()[genTypeName(obj, "NumericFieldsArg")]
		if !enumExists {
			fieldsEnumCfg := gql.EnumConfig{
				Name:   genTypeName(obj, "NumericFieldsArg"),
				Values: gql.EnumValueConfigMap{},
			}

			hasSumableFields := false
			// generate basic filter operator blocks for all the sumable types
			for _, field := range obj.Fields() {
				if field.Type == gql.Float || field.Type == gql.Int {
					hasSumableFields = true
					fieldsEnumCfg.Values[field.Name] = &gql.EnumValueConfig{Value: field.Name}
					continue
				}

				if list, isList := field.Type.(*gql.List); isList {
					hasSumableFields = true
					if isNumericArray(list) {
						fieldsEnumCfg.Values[field.Name] = &gql.EnumValueConfig{Value: field.Name}
					} else {
						// If it is a related list, we need to add count in here so that we can sum it
						fieldsEnumCfg.Values[parserTypes.CountFieldName] = &gql.EnumValueConfig{Value: parserTypes.CountFieldName}
					}
				}
			}
			// A child aggregate will always be aggregatable, as it can be present via an inner grouping
			fieldsEnumCfg.Values[parserTypes.SumFieldName] = &gql.EnumValueConfig{Value: parserTypes.SumFieldName}
			fieldsEnumCfg.Values[parserTypes.AverageFieldName] = &gql.EnumValueConfig{Value: parserTypes.AverageFieldName}

			if !hasSumableFields {
				return nil, nil
			}

			fieldsEnum = gql.NewEnum(fieldsEnumCfg)

			err := g.manager.schema.AppendType(fieldsEnum)
			if err != nil {
				return nil, err
			}
		}

		return gql.InputObjectConfigFieldMap{
			"field": &gql.InputObjectFieldConfig{
				Type: gql.NewNonNull(fieldsEnum),
			},
			parserTypes.LimitClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: "The maximum number of child items to aggregate.",
			},
			parserTypes.OffsetClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: "The index from which to start aggregating items.",
			},
			parserTypes.OrderClause: &gql.InputObjectFieldConfig{
				Type:        g.manager.schema.TypeMap()[genTypeName(obj, "OrderArg")],
				Description: "The order in which to aggregate items.",
			},
		}, nil
	}

	return gql.NewInputObject(gql.InputObjectConfig{
		Name:   genNumericObjectSelectorName(obj.Name()),
		Fields: fieldThunk,
	})
}

// Given a parsed ast.Node object, lookup the type in the TypeMap and return if its there
// otherwise return an error
// ast.Node, can either be a ast.Named type, a ast.List, or a ast.NonNull.
// The latter two are wrappers, and need to be further extracted
func astNodeToGqlType(typeMap map[string]gql.Type, t ast.Type) (gql.Type, error) {
	if t == nil {
		return nil, errors.New("type can't be nil")
	}

	switch astTypeVal := t.(type) {
	case *ast.List: // extract the underlying type and create a new
		// list instance of that type
		ttype, err := astNodeToGqlType(typeMap, astTypeVal.Type)
		if err != nil {
			return nil, err
		}

		return gql.NewList(ttype), nil

	case *ast.NonNull: // extract the underlying type and create a new
		// NonNull instance of that type
		ttype, err := astNodeToGqlType(typeMap, astTypeVal.Type)
		if err != nil {
			return nil, err
		}

		return gql.NewNonNull(ttype), nil
	}

	// default case, named type
	name := t.(*ast.Named).Name.Value
	ttype, ok := typeMap[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("No type found for given name: %s", name))
	}

	return ttype, nil
}

// GenerateQueryInputForGQLType is the main generation function
// for creating the full DefraDB Query schema for a given
// developer defined type
func (g *Generator) GenerateQueryInputForGQLType(
	ctx context.Context,
	obj *gql.Object,
) (*gql.Field, error) {
	if obj.Error() != nil {
		return nil, obj.Error()
	}
	types := queryInputTypeConfig{}
	types.filter = g.genTypeFilterArgInput(obj)

	// @todo: Don't add sub fields to filter/order for object list types
	types.groupBy = g.genTypeFieldsEnum(obj)
	types.order = g.genTypeOrderArgInput(obj)

	queryField := g.genTypeQueryableFieldList(ctx, obj, types)

	return queryField, nil
}

// GenerateMutationInputForGQLType creates all the mutation types and fields
// for the given graphQL object. It assumes that all the various
// filterArgs for the given type already exists, and will error otherwise.
func (g *Generator) GenerateMutationInputForGQLType(obj *gql.Object) ([]*gql.Field, error) {
	if obj.Error() != nil {
		return nil, obj.Error()
	}

	typeName := obj.Name()
	filter, ok := g.manager.schema.TypeMap()[typeName+"FilterArg"].(*gql.InputObject)
	if !ok {
		return nil, errors.New("Missing filter arg for mutation type generation " + typeName)
	}

	return g.genTypeMutationFields(obj, filter)
}

func (g *Generator) genTypeMutationFields(
	obj *gql.Object,
	filterInput *gql.InputObject,
) ([]*gql.Field, error) {
	create, err := g.genTypeMutationCreateField(obj)
	if err != nil {
		return nil, err
	}
	update, err := g.genTypeMutationUpdateField(obj, filterInput)
	if err != nil {
		return nil, err
	}
	delete, err := g.genTypeMutationDeleteField(obj, filterInput)
	if err != nil {
		return nil, err
	}
	return []*gql.Field{create, update, delete}, nil
}

func (g *Generator) genTypeMutationCreateField(obj *gql.Object) (*gql.Field, error) {
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: "create_" + obj.Name(),
		Type: obj,
		Args: gql.FieldConfigArgument{
			"data": schemaTypes.NewArgConfig(gql.String),
		},
	}
	return field, nil
}

func (g *Generator) genTypeMutationUpdateField(
	obj *gql.Object,
	filter *gql.InputObject,
) (*gql.Field, error) {
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: "update_" + obj.Name(),
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"id":     schemaTypes.NewArgConfig(gql.ID),
			"ids":    schemaTypes.NewArgConfig(gql.NewList(gql.ID)),
			"filter": schemaTypes.NewArgConfig(filter),
			"data":   schemaTypes.NewArgConfig(gql.String),
		},
	}
	return field, nil
}

func (g *Generator) genTypeMutationDeleteField(
	obj *gql.Object,
	filter *gql.InputObject,
) (*gql.Field, error) {
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: "delete_" + obj.Name(),
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"id":     schemaTypes.NewArgConfig(gql.ID),
			"ids":    schemaTypes.NewArgConfig(gql.NewList(gql.ID)),
			"filter": schemaTypes.NewArgConfig(filter),
			// "data":   newArgConfig(gql.String),
		},
	}
	return field, nil
}

// enum {Type.Name}Fields { ... }
func (g *Generator) genTypeFieldsEnum(obj *gql.Object) *gql.Enum {
	enumFieldsCfg := gql.EnumConfig{
		Name:   genTypeName(obj, "Fields"),
		Values: gql.EnumValueConfigMap{},
	}

	for f, field := range obj.Fields() {
		enumFieldsCfg.Values[field.Name] = &gql.EnumValueConfig{Value: f}
	}

	return gql.NewEnum(enumFieldsCfg)
}

// input {Type.Name}FilterArg { ... }
func (g *Generator) genTypeFilterArgInput(obj *gql.Object) *gql.InputObject {
	var selfRefType *gql.InputObject

	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterArg"),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(
		func() (gql.InputObjectConfigFieldMap, error) {
			fields := gql.InputObjectConfigFieldMap{}

			// conditionals
			compoundListType := &gql.InputObjectFieldConfig{
				Type: gql.NewList(selfRefType),
			}

			fields["_and"] = compoundListType
			fields["_or"] = compoundListType
			fields["_not"] = &gql.InputObjectFieldConfig{
				Type: selfRefType,
			}

			// generate basic filter operator blocks
			// @todo: Extract object field loop into its own utility func
			for f, field := range obj.Fields() {
				if _, ok := parserTypes.ReservedFields[f]; ok && f != parserTypes.DocKeyFieldName {
					continue
				}
				// scalars (leafs)
				if gql.IsLeafType(field.Type) {
					if _, isList := field.Type.(*gql.List); isList {
						// Filtering by inline array value is currently not supported
						continue
					}
					operatorType, isFilterable := g.manager.schema.TypeMap()[field.Type.Name()+"OperatorBlock"]
					if !isFilterable {
						continue
					}
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: operatorType,
					}
				} else { // objects (relations)
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: g.manager.schema.TypeMap()[genTypeName(field.Type, "FilterArg")],
					}
				}
			}

			return fields, nil
		},
	)

	// add the fields thunker
	inputCfg.Fields = fieldThunk
	selfRefType = gql.NewInputObject(inputCfg)
	return selfRefType
}

func (g *Generator) genLeafFilterArgInput(obj gql.Type) *gql.InputObject {
	var selfRefType *gql.InputObject

	var filterTypeName string
	if notNull, isNotNull := obj.(*gql.NonNull); isNotNull {
		// GQL does not support '!' in type names, and so we have to manipulate the
		// underlying name like this if it is a nullable type.
		filterTypeName = fmt.Sprintf("NotNull%s", notNull.OfType.Name())
	} else {
		filterTypeName = obj.Name()
	}

	inputCfg := gql.InputObjectConfig{
		Name: fmt.Sprintf("%s%s", filterTypeName, "FilterArg"),
	}

	var fieldThunk gql.InputObjectConfigFieldMapThunk = func() (gql.InputObjectConfigFieldMap, error) {
		fields := gql.InputObjectConfigFieldMap{}

		compoundListType := &gql.InputObjectFieldConfig{
			Type: gql.NewList(selfRefType),
		}

		fields["_and"] = compoundListType
		fields["_or"] = compoundListType

		operatorBlockName := fmt.Sprintf("%s%s", filterTypeName, "OperatorBlock")
		operatorType, hasOperatorType := g.manager.schema.TypeMap()[operatorBlockName]
		if !hasOperatorType {
			// This should be impossible
			return nil, errors.New("Operator block not found", errors.NewKV("Name", operatorBlockName))
		}

		operatorObject, isInputObj := operatorType.(*gql.InputObject)
		if !isInputObj {
			// This should be impossible
			return nil, errors.New(
				"Invalid cast",
				errors.NewKV("Expected type", "*gql.InputObject"),
				errors.NewKV("Actual type", fmt.Sprintf("%T", operatorType)),
			)
		}

		for f, field := range operatorObject.Fields() {
			fields[f] = &gql.InputObjectFieldConfig{
				Type: field.Type,
			}
		}

		return fields, nil
	}

	inputCfg.Fields = fieldThunk
	selfRefType = gql.NewInputObject(inputCfg)
	return selfRefType
}

func (g *Generator) genTypeOrderArgInput(obj *gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "OrderArg"),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(
		func() (gql.InputObjectConfigFieldMap, error) {
			fields := gql.InputObjectConfigFieldMap{}

			for f, field := range obj.Fields() {
				if _, ok := parserTypes.ReservedFields[f]; ok && f != parserTypes.DocKeyFieldName {
					continue
				}
				typeMap := g.manager.schema.TypeMap()
				if gql.IsLeafType(field.Type) { // only Scalars, and enums
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: typeMap["Ordering"],
					}
				} else { // sub objects
					configType, isOrderable := typeMap[genTypeName(field.Type, "OrderArg")]
					if !isOrderable {
						fields[field.Name] = &gql.InputObjectFieldConfig{
							Type: &gql.InputObjectField{},
						}
					} else {
						fields[field.Name] = &gql.InputObjectFieldConfig{
							Type: configType,
						}
					}
				}
			}

			return fields, nil
		},
	)

	inputCfg.Fields = fieldThunk
	return gql.NewInputObject(inputCfg)
}

type queryInputTypeConfig struct {
	filter  *gql.InputObject
	groupBy *gql.Enum
	order   *gql.InputObject
}

func (g *Generator) genTypeQueryableFieldList(
	ctx context.Context,
	obj *gql.Object,
	config queryInputTypeConfig,
) *gql.Field {
	name := obj.Name()

	// add the generated types to the type map
	if err := g.manager.schema.AppendType(config.filter); err != nil {
		log.ErrorE(
			ctx,
			"Failed to append runtime schema with filter",
			err,
			logging.NewKV("SchemaItem", config.filter),
		)
	}

	if err := g.manager.schema.AppendType(config.groupBy); err != nil {
		log.ErrorE(
			ctx,
			"Failed to append runtime schema with groupBy",
			err,
			logging.NewKV("SchemaItem", config.groupBy),
		)
	}

	if err := g.manager.schema.AppendType(config.order); err != nil {
		log.ErrorE(
			ctx,
			"Failed to append runtime schema with order",
			err,
			logging.NewKV("SchemaItem", config.order),
		)
	}

	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: name,
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"dockey":                 schemaTypes.NewArgConfig(gql.String),
			"dockeys":                schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(gql.String))),
			"cid":                    schemaTypes.NewArgConfig(gql.String),
			"filter":                 schemaTypes.NewArgConfig(config.filter),
			"groupBy":                schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(config.groupBy))),
			"order":                  schemaTypes.NewArgConfig(config.order),
			parserTypes.LimitClause:  schemaTypes.NewArgConfig(gql.Int),
			parserTypes.OffsetClause: schemaTypes.NewArgConfig(gql.Int),
		},
	}

	return field
}

func (g *Generator) appendIfNotExists(obj gql.Type) error {
	if _, typeExists := g.manager.schema.TypeMap()[obj.Name()]; !typeExists {
		err := g.manager.schema.AppendType(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// Reset the stateful data within a Generator.
// Usually called after a round of type generation
func (g *Generator) Reset() {
	g.typeDefs = make([]*gql.Object, 0)
	g.expandedFields = make(map[string]bool)
}

func genTypeName(obj gql.Type, name string) string {
	return fmt.Sprintf("%s%s", obj.Name(), name)
}

// isNumericArray returns true if the given list is a list of numerical values.
func isNumericArray(list *gql.List) bool {
	// We have to compare the names here, as the gql lib we use
	// does not have an easier way to compare non-nullable types
	return list.OfType.Name() == gql.NewNonNull(gql.Float).Name() ||
		list.OfType.Name() == gql.NewNonNull(gql.Int).Name() ||
		list.OfType == gql.Int ||
		list.OfType == gql.Float
}

// find a given directive
func findDirective(field *ast.FieldDefinition, directiveName string) (*ast.Directive, bool) {
	for _, directive := range field.Directives {
		if directive.Name.Value == directiveName {
			return directive, true
		}
	}
	return nil, false
}

/* Example

typeDefs := ` ... `

ast, err := parserTypes.Parse(typeDefs)
types, err := buildTypesFromAST(ast)

types, err := GenerateDBQuerySchema(ast)
schemaManager.Update(types)

// request
q := query.Parse(qry)
qplan := planner.Plan(q, schemaManager.Schema)
resp := db.queryEngine.Execute(ctx, q, qplan)


*/
