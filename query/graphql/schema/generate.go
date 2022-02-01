// Copyright 2020 Source Inc.
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
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema/types"

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
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
func (g *Generator) FromSDL(schema string) ([]*gql.Object, *ast.Document, error) {
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
	types, err := g.FromAST(doc)
	return types, doc, err
}

func (g *Generator) FromAST(document *ast.Document) ([]*gql.Object, error) {
	typeMapBeforeMutation := g.manager.schema.TypeMap()
	typesBeforeMutation := make(map[string]interface{}, len(typeMapBeforeMutation))

	for typeName := range typeMapBeforeMutation {
		typesBeforeMutation[typeName] = struct{}{}
	}

	result, err := g.fromAST(document)

	if err != nil {
		// If there is an error we should drop any new objects as they may be partial, poluting the in memory cache
		// This is quite a simple check at the moment (on type name) - this should be expanded when we allow schema mutation/deletion
		// There is no guarantee that `typeMapBeforeMutation` will still be the object returned by `schema.TypeMap()`, so we should re-fetch it
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
func (g *Generator) fromAST(document *ast.Document) ([]*gql.Object, error) {
	// build base types
	defs, err := g.buildTypesFromAST(document)
	if err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	if err := g.genAggregateFields(); err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	generatedFilterBaseArgs := make([]*gql.InputObject, len(g.typeDefs))
	for i, t := range g.typeDefs {
		generatedFilterBaseArgs[i] = g.genTypeFilterBaseArgInput(t)
	}

	for _, t := range generatedFilterBaseArgs {
		err := g.manager.schema.AppendType(t)
		if err != nil {
			// Todo: better error handle
			log.Printf("failure appending type while generating query type defs from an AST : %v", err)
		}
	}

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// for each built type
	// 		generate query inputs
	queryType := g.manager.schema.QueryType()
	generatedQueryFields := make([]*gql.Field, 0)
	for _, t := range g.typeDefs {
		f, err := g.GenerateQueryInputForGQLType(t)
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

	// secondary pass to expand query collection type
	// argument inputs
	// query := g.manager.schema.QueryType()
	// queries := query.Fields()
	// only apply to generated query fields, and only once
	for _, def := range generatedQueryFields {
		t := def.Type
		if obj, ok := t.(*gql.List); ok {
			if err := g.expandInputArgument(obj.OfType.(*gql.Object)); err != nil {
				return nil, err
			}
		}
	}

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// now lets generate the mutation types.
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
		// ignore reserved fields, execpt the Group field (as that requires typing)
		if _, ok := parser.ReservedFields[f]; ok && f != parser.GroupFieldName {
			continue
		}
		// Both the object name and the field name should be used as the key
		// in case the child object type is referenced multiple times from the same parent type
		fieldKey := obj.Name() + f
		switch t := def.Type.(type) {
		case *gql.Object:
			if _, complete := g.expandedFields[fieldKey]; complete {
				continue
			} else {
				g.expandedFields[fieldKey] = true
			}
			// make sure all the sub fields are expanded first
			if err := g.expandInputArgument(t); err != nil {
				return err
			}

			// new field object with arguments (single)
			expandedField, err := g.createExpandedFieldSingle(def, t)
			if err != nil {
				return err
			}

			// obj.AddFieldConfig(f, expandedField)
			// obj := g.manager.schema.Type(obj.Name()).(*gql.Object)
			obj.AddFieldConfig(f, expandedField)

		case *gql.List: // new field object with arguments (list)
			listType := t.OfType
			if _, complete := g.expandedFields[fieldKey]; complete {
				continue
			} else {
				g.expandedFields[fieldKey] = true
			}

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
			// @todo: check if NonNull is possible here
			//case *gql.NonNull:
			// get subtype
		}
	}

	return nil
}

func (g *Generator) createExpandedFieldSingle(f *gql.FieldDefinition, t *gql.Object) (*gql.Field, error) {
	typeName := t.Name()
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: f.Name,
		Type: t,
		Args: gql.FieldConfigArgument{
			"filter": newArgConfig(g.manager.schema.TypeMap()[typeName+"FilterArg"]),
		},
	}
	return field, nil
}

// @todo: add field reference so we can copy extra fields (like description, depreciation, etc)
func (g *Generator) createExpandedFieldList(f *gql.FieldDefinition, t *gql.Object) (*gql.Field, error) {
	typeName := t.Name()
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: f.Name,
		Type: gql.NewList(t),
		Args: gql.FieldConfigArgument{
			"filter":  newArgConfig(g.manager.schema.TypeMap()[typeName+"FilterArg"]),
			"groupBy": newArgConfig(gql.NewList(gql.NewNonNull(g.manager.schema.TypeMap()[typeName+"Fields"]))),
			"having":  newArgConfig(g.manager.schema.TypeMap()[typeName+"HavingArg"]),
			"order":   newArgConfig(g.manager.schema.TypeMap()[typeName+"OrderArg"]),
			"limit":   newArgConfig(gql.Int),
			"offset":  newArgConfig(gql.Int),
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
func (g *Generator) buildTypesFromAST(document *ast.Document) ([]*gql.Object, error) {
	// @todo: Check for duplicate named defined types in the TypeMap
	// get all the defined types from the AST
	objs := make([]*gql.Object, 0)

	for _, def := range document.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			// check if type exists
			if _, ok := g.manager.schema.TypeMap()[defType.Name.Value]; ok {
				return nil, fmt.Errorf("Schema type already exists: %s", defType.Name.Value)
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
				fields["_key"] = &gql.Field{Type: gql.ID}

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

						_, err = g.manager.Relations.RegisterSingle(relName, ttype.Name(), fType.Name, base.Meta_Relation_ONE)
						if err != nil {
							// Todo: better error handle
							log.Printf("got error while registering single relation: %v", err)
						}

					case *gql.List:
						ltype := subobj.OfType
						// register the relation
						relName, err := getRelationshipName(field, objconf, ltype)
						if err != nil {
							return nil, err
						}

						_, err = g.manager.Relations.RegisterSingle(relName, ltype.Name(), fType.Name, base.Meta_Relation_MANY)
						if err != nil {
							// Todo: better error handle
							log.Printf("got error while registering single relation: %v", err)
						}
					}

					fType.Type = ttype
					fields[fType.Name] = fType
				}

				// add _version field
				fields["_version"] = &gql.Field{
					Type: gql.NewList(types.Commit),
				}

				// @todo Pairup on removing the staticcheck linter error below.
				gqlType, ok := g.manager.schema.TypeMap()[defType.Name.Value]
				if !ok {
					return nil, fmt.Errorf("object not found whilst executing fields thunk: %s", defType.Name.Value)
				}

				fields[parser.GroupFieldName] = &gql.Field{
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
func getRelationshipName(field *ast.FieldDefinition, hostName gql.ObjectConfig, targetName gql.Type) (string, error) {
	// search for a user-defined name, and return it if found
	for _, directive := range field.Directives {
		if directive.Name.Value == "relation" {
			for _, arguement := range directive.Arguments {
				if arguement.Name.Value == "name" {
					name, isString := arguement.Value.GetValue().(string)
					if !isString {
						return "", fmt.Errorf("Relationship name must be of type string, but was: %v", arguement.Value.GetKind())
					}
					return name, nil
				}
			}
		}
	}

	// if no name is provided, generate one
	return genRelationName(hostName.Name, targetName.Name())
}

func (g *Generator) genAggregateFields() error {
	for _, t := range g.typeDefs {
		countField, err := g.genCountFieldConfig(t)
		if err != nil {
			return err
		}
		t.AddFieldConfig(countField.Name, &countField)

		sumField := g.genSumFieldConfig(t)
		t.AddFieldConfig(sumField.Name, &sumField)
	}

	return nil
}

func (g *Generator) genCountFieldConfig(obj *gql.Object) (gql.Field, error) {
	inputCfg := gql.EnumConfig{
		Name:   genTypeName(obj, "CountArg"),
		Values: gql.EnumValueConfigMap{},
	}

	for _, field := range obj.Fields() {
		// Only lists can be counted
		if _, isList := field.Type.(*gql.List); !isList {
			continue
		}
		inputCfg.Values[field.Name] = &gql.EnumValueConfig{Value: field.Name}
	}
	countType := gql.NewEnum(inputCfg)
	err := g.manager.schema.AppendType(countType)
	if err != nil {
		return gql.Field{}, err
	}

	field := gql.Field{
		Name: parser.CountFieldName,
		Type: gql.Int,
		Args: gql.FieldConfigArgument{
			"field": newArgConfig(countType),
		},
	}

	return field, nil
}

func (g *Generator) genSumFieldConfig(obj *gql.Object) gql.Field {
	var sumType *gql.InputObject

	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "SumArg"),
	}

	inputCfg.Fields = (gql.InputObjectConfigFieldMapThunk)(func() (gql.InputObjectConfigFieldMap, error) {
		fields := gql.InputObjectConfigFieldMap{}

		sumBaseArgType, isSumable := g.genSumBaseArgInput(obj)
		if isSumable {
			err := g.manager.schema.AppendType(sumBaseArgType)
			if err != nil {
				return gql.InputObjectConfigFieldMap{}, err
			}
		}

		for _, field := range obj.Fields() {
			// we can only sum list items
			listType, isList := field.Type.(*gql.List)
			if !isList {
				continue
			}

			if listType.OfType == gql.Float || listType.OfType == gql.Int {
				// If it is an inline scalar array then we require an empty object as an argument due to the lack of union input types
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: &gql.Object{},
				}
			} else {
				subSumType, isSubTypeSumable := g.manager.schema.TypeMap()[genTypeName(field.Type, "SumBaseArg")]
				// If the item is not in the type map, it must contain no summable fields (e.g. no Int/Floats)
				if !isSubTypeSumable {
					continue
				}
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: subSumType,
				}
			}

		}

		return fields, nil
	})
	sumType = gql.NewInputObject(inputCfg)

	//this might resolve the thunk?  Race issue?
	err := g.manager.schema.AppendType(sumType)
	if err != nil {
		log.Printf("failure appending sumType : %v", err)
	}

	field := gql.Field{
		Name: parser.SumFieldName,
		Type: gql.Float,
		Args: gql.FieldConfigArgument{
			"field": newArgConfig(sumType),
		},
	}
	return field
}

func (g *Generator) genSumBaseArgInput(obj *gql.Object) (*gql.Enum, bool) {
	inputCfg := gql.EnumConfig{
		Name:   genTypeName(obj, "SumBaseArg"),
		Values: gql.EnumValueConfigMap{},
	}

	hasSumableFields := false
	// generate basic filter operator blocks for all the sumable types
	for _, field := range obj.Fields() {
		if field.Type == gql.Float || field.Type == gql.Int {
			hasSumableFields = true
			inputCfg.Values[field.Name] = &gql.EnumValueConfig{Value: field.Name}
		}
	}

	if !hasSumableFields {
		return nil, false
	}

	return gql.NewEnum(inputCfg), true
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
		return nil, fmt.Errorf("No type found for given name: %s", name)
	}

	return ttype, nil
}

// type SchemaObject

// GenerateQueryInputForGQLType is the main generation function
// for creating the full DefraDB Query schema for a given
// developer defined type
func (g *Generator) GenerateQueryInputForGQLType(obj *gql.Object) (*gql.Field, error) {
	if obj.Error() != nil {
		return nil, obj.Error()
	}
	types := queryInputTypeConfig{}
	types.filter = g.genTypeFilterArgInput(obj)

	// @todo: Don't add sub fields to filter/order for object list types
	types.groupBy = g.genTypeFieldsEnum(obj)
	types.having = g.genTypeHavingArgInput(obj)
	types.order = g.genTypeOrderArgInput(obj)
	// var queryField *gql.Field
	queryField := g.genTypeQueryableFieldList(obj, types)

	// queryType := g.manager.schema.QueryType()
	// queryType.AddFieldConfig(queryField.Name, queryField)

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

func (g *Generator) genTypeMutationFields(obj *gql.Object, filterInput *gql.InputObject) ([]*gql.Field, error) {
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
			"data": newArgConfig(gql.String),
		},
	}
	return field, nil
}

func (g *Generator) genTypeMutationUpdateField(obj *gql.Object, filter *gql.InputObject) (*gql.Field, error) {
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: "update_" + obj.Name(),
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"id":     newArgConfig(gql.ID),
			"ids":    newArgConfig(gql.NewList(gql.ID)),
			"filter": newArgConfig(filter),
			"data":   newArgConfig(gql.String),
		},
	}
	return field, nil
}

func (g *Generator) genTypeMutationDeleteField(obj *gql.Object, filter *gql.InputObject) (*gql.Field, error) {
	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: "delete_" + obj.Name(),
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"id":     newArgConfig(gql.ID),
			"filter": newArgConfig(filter),
			"data":   newArgConfig(gql.String),
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
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(func() (gql.InputObjectConfigFieldMap, error) {
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
			if _, ok := parser.ReservedFields[f]; ok && f != "_key" {
				continue
			}
			// scalars (leafs)
			if gql.IsLeafType(field.Type) { // only Scalars, and enums
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: g.manager.schema.TypeMap()[genTypeName(field.Type, "OperatorBlock")],
				}
			} else { // objects (relations)
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: g.manager.schema.TypeMap()[genTypeName(field.Type, "FilterBaseArg")],
				}
			}

		}

		// fmt.Println("#####################")
		// spew.Dump(fields)
		return fields, nil
	})

	// add the fields thunker
	inputCfg.Fields = fieldThunk
	selfRefType = gql.NewInputObject(inputCfg)
	return selfRefType
}

// input {Type.Name}FilterBaseArg { ... }
func (g *Generator) genTypeFilterBaseArgInput(obj *gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterBaseArg"),
	}
	fields := gql.InputObjectConfigFieldMap{}
	// generate basic filter operator blocks for all the Leaf types
	// (scalars + enums)
	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: g.manager.schema.TypeMap()[field.Type.Name()+"OperatorBlock"],
			}
		}
	}

	inputCfg.Fields = fields
	return gql.NewInputObject(inputCfg)
}

// query spec - sec N
func (g *Generator) genTypeHavingArgInput(obj *gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "HavingArg"),
	}
	fields := gql.InputObjectConfigFieldMap{}
	havingBlock := g.genTypeHavingBlockInput(obj)

	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: havingBlock,
			}
		}
	}

	inputCfg.Fields = fields
	return gql.NewInputObject(inputCfg)
}

func (g *Generator) genTypeHavingBlockInput(obj *gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "HavingBlock"),
	}
	fields := gql.InputObjectConfigFieldMap{}

	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: g.manager.schema.TypeMap()["FloatOperatorBlock"],
			}
		}
	}

	inputCfg.Fields = fields
	return gql.NewInputObject(inputCfg)
}

func (g *Generator) genTypeOrderArgInput(obj *gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "OrderArg"),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(func() (gql.InputObjectConfigFieldMap, error) {
		fields := gql.InputObjectConfigFieldMap{}

		for f, field := range obj.Fields() {
			if _, ok := parser.ReservedFields[f]; ok && f != "_key" {
				continue
			}
			if gql.IsLeafType(field.Type) { // only Scalars, and enums
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: g.manager.schema.TypeMap()["Ordering"],
				}
			} else { // sub objects
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: g.manager.schema.TypeMap()[genTypeName(field.Type, "OrderArg")],
				}
			}
		}

		return fields, nil
	})

	inputCfg.Fields = fieldThunk
	return gql.NewInputObject(inputCfg)
}

type queryInputTypeConfig struct {
	filter  *gql.InputObject
	groupBy *gql.Enum
	having  *gql.InputObject
	order   *gql.InputObject
}

func (g *Generator) genTypeQueryableFieldList(obj *gql.Object, config queryInputTypeConfig) *gql.Field {
	name := strings.ToLower(obj.Name())

	// add the generated types to the type map
	if err := g.manager.schema.AppendType(config.filter); err != nil {
		// Todo: better error handle
		log.Printf("failure appending runtime schema - filter: %v", err)
	}

	if err := g.manager.schema.AppendType(config.groupBy); err != nil {
		// Todo: better error handle
		log.Printf("failure appending runtime schema - groupBy: %v", err)
	}

	if err := g.manager.schema.AppendType(config.having); err != nil {
		// Todo: better error handle
		log.Printf("failure appending runtime schema - having: %v", err)
	}

	if err := g.manager.schema.AppendType(config.order); err != nil {
		// Todo: better error handle
		log.Printf("failure appending runtime schema - order: %v", err)
	}

	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: name,
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			"dockey":  newArgConfig(gql.String),
			"dockeys": newArgConfig(gql.NewList(gql.NewNonNull(gql.String))),
			"cid":     newArgConfig(gql.String),
			"filter":  newArgConfig(config.filter),
			"groupBy": newArgConfig(gql.NewList(gql.NewNonNull(config.groupBy))),
			"having":  newArgConfig(config.having),
			"order":   newArgConfig(config.order),
			"limit":   newArgConfig(gql.Int),
			"offset":  newArgConfig(gql.Int),
		},
	}

	return field
}

// Reset the stateful data within a Generator.
// Usually called after a round of type generation
func (g *Generator) Reset() {
	g.typeDefs = make([]*gql.Object, 0)
	g.expandedFields = make(map[string]bool)
}

func newArgConfig(t gql.Input) *gql.ArgumentConfig {
	return &gql.ArgumentConfig{
		Type: t,
	}
}

func genTypeName(obj gql.Type, name string) string {
	return fmt.Sprintf("%s%s", obj.Name(), name)
}

/* Example

typeDefs := ` ... `

ast, err := parser.Parse(typeDefs)
types, err := buildTypesFromAST(ast)

types, err := GenerateDBQuerySchema(ast)
schemaManager.Update(types)

// request
q := query.Parse(qry)
qplan := planner.Plan(q, schemaManager.Schema)
resp := db.queryEngine.Execute(ctx, q, qplan)


*/
