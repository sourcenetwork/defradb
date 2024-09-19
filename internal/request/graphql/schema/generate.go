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
	"strings"

	gql "github.com/sourcenetwork/graphql-go"

	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/client/request"
	schemaTypes "github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
)

// Given a basic developer defined schema in GraphQL Schema Definition Language
// create a fully DefraDB complaint GraphQL schema using a "code-first" dynamic
// approach

const (
	filterInputNameSuffix    = "FilterArg"
	mutationInputNameSuffix  = "MutationInputArg"
	mutationInputsNameSuffix = "MutationInputsArg"
)

const (
	typeFieldEnumSuffix         = "Field"
	typeExplicitFieldEnumSuffix = "ExplicitField"
)

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

// Generate generates the query-op and mutation-op type definitions from
// the given CollectionDescriptions.
func (g *Generator) Generate(ctx context.Context, collections []client.CollectionDefinition) ([]*gql.Object, error) {
	typeMapBeforeMutation := g.manager.schema.TypeMap()
	typesBeforeMutation := make(map[string]any, len(typeMapBeforeMutation))

	for typeName := range typeMapBeforeMutation {
		typesBeforeMutation[typeName] = struct{}{}
	}

	result, err := g.generate(ctx, collections)

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

// generate generates the query-op and mutation-op type definitions from
// the given CollectionDescriptions.
func (g *Generator) generate(ctx context.Context, collections []client.CollectionDefinition) ([]*gql.Object, error) {
	// build base types
	defs, err := g.buildTypes(collections)
	if err != nil {
		return nil, err
	}
	// build mutation input types
	err = g.buildMutationInputTypes(collections)
	if err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// for each built type generate query inputs
	queryType := g.manager.schema.QueryType()
	subscriptionType := g.manager.schema.SubscriptionType()
	generatedQueryFields := make([]*gql.Field, 0)
	for _, t := range g.typeDefs {
		f, err := g.GenerateQueryInputForGQLType(ctx, t)
		if err != nil {
			return nil, err
		}
		generatedQueryFields = append(generatedQueryFields, f)

		var isEmbedded bool
		for _, definition := range collections {
			if t.Name() == definition.Schema.Name && !definition.Description.Name.HasValue() {
				isEmbedded = true
				break
			}
		}

		// If the object is embedded, it may not be queried directly, so we must not add it
		// to the `query` object.  We do however need the query-input objects to be generated
		// (further up in this block), as they are still required for stuff like grouping.
		if isEmbedded {
			continue
		}

		queryType.AddFieldConfig(f.Name, f)
		subscriptionType.AddFieldConfig(f.Name, f)
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
			if _, isAggregate := request.Aggregates[def.Name]; isAggregate {
				for name, aggregateTarget := range def.Args {
					expandedField := &gql.InputObjectFieldConfig{
						Description: aggregateFilterArgDescription,
						Type:        g.manager.schema.TypeMap()[name+filterInputNameSuffix],
					}
					aggregateTarget.Type.(*gql.InputObject).AddFieldConfig(request.FilterClause, expandedField)
				}
			}
		}
	}

	g.appendCommitChildGroupField()

	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// now let's generate the mutation types.
	mutationType := g.manager.schema.MutationType()
	for _, t := range g.typeDefs {
		// Note: Whilst the `isReadOnly` code is fairly unpleasent, it will hopefully not live for too much longer
		// as we plan to transition to DQL.
		var isReadOnly bool
		var collectionFound bool
		for _, definition := range collections {
			if t.Name() == definition.Description.Name.Value() {
				isReadOnly = len(definition.Description.QuerySources()) > 0
				collectionFound = true
				break
			}
		}
		if !collectionFound {
			// If we did not find a collection with this name, check for matching schemas (embedded objects)
			for _, definition := range collections {
				if t.Name() == definition.Schema.Name {
					// All embedded objects are readonly
					isReadOnly = true
					collectionFound = true
					break
				}
			}
		}

		if isReadOnly {
			// We do not currently allow mutation via views, so don't add them to the mutation object
			continue
		}

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
		if _, ok := request.ReservedFields[f]; ok && f != request.GroupFieldName {
			if _, isAggregate := request.Aggregates[f]; !isAggregate {
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
			if _, isAggregate := request.Aggregates[f]; isAggregate {
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
		if target == request.GroupFieldName {
			filterTypeName = obj.Name() + filterInputNameSuffix
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
						filterTypeName = genTypeName(list.OfType, filterInputNameSuffix)
					}
				} else {
					filterTypeName = targeted.Type.Name() + filterInputNameSuffix
				}
			} else {
				return NewErrAggregateTargetNotFound(obj.Name(), target)
			}
		}

		if filterType, canHaveFilter := g.manager.schema.TypeMap()[filterTypeName]; canHaveFilter {
			// Sometimes a filter is not permitted, for example when aggregating `_version`
			expandedField := &gql.InputObjectFieldConfig{
				Description: aggregateFilterArgDescription,
				Type:        filterType,
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
		Name:        f.Name,
		Description: f.Description,
		Type:        t,
		Args: gql.FieldConfigArgument{
			"filter": schemaTypes.NewArgConfig(
				g.manager.schema.TypeMap()[typeName+filterInputNameSuffix],
				singleFieldFilterArgDescription,
			),
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
		Name:        f.Name,
		Description: f.Description,
		Type:        gql.NewList(t),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName:  schemaTypes.NewArgConfig(gql.String, docIDArgDescription),
			request.DocIDsArgName: schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(gql.String)), docIDsArgDescription),
			"filter": schemaTypes.NewArgConfig(
				g.manager.schema.TypeMap()[typeName+filterInputNameSuffix],
				listFieldFilterArgDescription,
			),
			"groupBy": schemaTypes.NewArgConfig(
				gql.NewList(gql.NewNonNull(g.manager.schema.TypeMap()[typeName+typeFieldEnumSuffix])),
				schemaTypes.GroupByArgDescription,
			),
			"order": schemaTypes.NewArgConfig(
				g.manager.schema.TypeMap()[typeName+"OrderArg"],
				schemaTypes.OrderArgDescription,
			),
			request.LimitClause:  schemaTypes.NewArgConfig(gql.Int, schemaTypes.LimitArgDescription),
			request.OffsetClause: schemaTypes.NewArgConfig(gql.Int, schemaTypes.OffsetArgDescription),
		},
	}

	return field, nil
}

// @todo: Add Schema Directives (IE: relation, etc..)

// @todo: Add validation support for the AST
// @body: Type generation is only supported for Object type definitions.
// Unions, Interfaces, etc are not currently supported.

// Given a set of developer defined collection types
// extract and return the correct gql.Object type(s)
func (g *Generator) buildTypes(
	collections []client.CollectionDefinition,
) ([]*gql.Object, error) {
	definitionCache := client.NewDefinitionCache(collections)

	// @todo: Check for duplicate named defined types in the TypeMap
	// get all the defined types from the AST
	objs := make([]*gql.Object, 0)

	for _, collection := range collections {
		fieldDescriptions := collection.GetFields()
		isEmbeddedObject := !collection.Description.Name.HasValue()
		isQuerySource := len(collection.Description.QuerySources()) > 0
		isViewObject := isEmbeddedObject || isQuerySource

		var objectName string
		if isEmbeddedObject {
			// If this is an embedded object, take the type name from the Schema
			objectName = collection.Schema.Name
		} else {
			objectName = collection.Description.Name.Value()
		}

		// check if type exists
		if _, ok := g.manager.schema.TypeMap()[objectName]; ok {
			return nil, NewErrSchemaTypeAlreadyExist(objectName)
		}

		objconf := gql.ObjectConfig{
			Name: objectName,
		}

		// Wrap field definition in a thunk so we can
		// handle any embedded object which is defined
		// at a future point in time.
		fieldsThunk := (gql.FieldsThunk)(func() (gql.Fields, error) {
			fields := gql.Fields{}

			if !isViewObject {
				// automatically add the _docID: ID field to the type
				fields[request.DocIDFieldName] = &gql.Field{
					Description: docIDFieldDescription,
					Type:        gql.ID,
				}
			}

			for _, field := range fieldDescriptions {
				if field.Name == request.DocIDFieldName {
					// The `_docID` field is included in the fieldDescriptions,
					// but we do not wish to override the standard definition
					// with the collection held definition (particularly the
					// description)
					continue
				}

				var ttype gql.Type
				if otherDef, ok := client.GetDefinition(definitionCache, collection, field.Kind); ok {
					ttype, ok = g.manager.schema.TypeMap()[otherDef.GetName()]
					if !ok {
						return nil, NewErrTypeNotFound(field.Kind.String())
					}
					if field.Kind.IsArray() {
						ttype = gql.NewList(ttype)
					}
				} else {
					var ok bool
					ttype, ok = fieldKindToGQLType[field.Kind]
					if !ok {
						return nil, NewErrTypeNotFound(field.Kind.String())
					}
				}

				fields[field.Name] = &gql.Field{
					Name: field.Name,
					Type: ttype,
				}
			}

			gqlType, ok := g.manager.schema.TypeMap()[objectName]
			if !ok {
				return nil, NewErrObjectNotFoundDuringThunk(objectName)
			}

			fields[request.GroupFieldName] = &gql.Field{
				Description: groupFieldDescription,
				Type:        gql.NewList(gqlType),
			}

			if !isViewObject {
				// add _version field
				fields[request.VersionFieldName] = &gql.Field{
					Description: versionFieldDescription,
					Type:        gql.NewList(g.manager.schema.TypeMap()[request.CommitTypeName]),
				}

				// add _deleted field
				fields[request.DeletedFieldName] = &gql.Field{
					Description: deletedFieldDescription,
					Type:        gql.Boolean,
				}
			}

			return fields, nil
		})

		objconf.Fields = fieldsThunk

		obj := gql.NewObject(objconf)
		objs = append(objs, obj)

		g.manager.schema.TypeMap()[obj.Name()] = obj
		g.typeDefs = append(g.typeDefs, obj)
	}

	return objs, nil
}

// buildMutationInputTypes creates the input object types
// for collection create and update mutation operations.
func (g *Generator) buildMutationInputTypes(collections []client.CollectionDefinition) error {
	for _, collection := range collections {
		if !collection.Description.Name.HasValue() {
			// If the definition's collection is empty, this must be a collectionless
			// schema, in which case users cannot mutate documents through it and we
			// have no need to build mutation input types for it.
			continue
		}

		mutationInputName := collection.Description.Name.Value() + mutationInputNameSuffix

		// check if mutation input type exists
		if _, ok := g.manager.schema.TypeMap()[mutationInputName]; ok {
			return NewErrMutationInputTypeAlreadyExist(mutationInputName)
		}

		mutationObjConf := gql.InputObjectConfig{
			Name: mutationInputName,
		}

		// Wrap mutation input object definition in a thunk so we can
		// handle any embedded object which is defined
		// at a future point in time.
		mutationObjConf.Fields = (gql.InputObjectConfigFieldMapThunk)(func() (gql.InputObjectConfigFieldMap, error) {
			fields := make(gql.InputObjectConfigFieldMap)

			for _, field := range collection.GetFields() {
				if strings.HasPrefix(field.Name, "_") {
					// ignore system defined args as the
					// user cannot override their values
					continue
				}

				var ttype gql.Type
				if field.Kind.IsObject() {
					if field.Kind.IsArray() {
						ttype = gql.NewList(gql.ID)
					} else {
						ttype = gql.ID
					}
				} else {
					var ok bool
					ttype, ok = fieldKindToGQLType[field.Kind]
					if !ok {
						return nil, NewErrTypeNotFound(fmt.Sprint(field.Kind))
					}
				}

				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: ttype,
				}
			}

			return fields, nil
		})

		mutationObj := gql.NewInputObject(mutationObjConf)
		g.manager.schema.TypeMap()[mutationObj.Name()] = mutationObj
	}

	return nil
}

func (g *Generator) genAggregateFields() error {
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
		Name:        request.CountFieldName,
		Description: schemaTypes.CountFieldDescription,
		Type:        gql.Int,
		Args:        gql.FieldConfigArgument{},
	}

	for name, inputObject := range topLevelCountInputs {
		topLevelCountField.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
	}

	return &topLevelCountField
}

func genTopLevelNumericAggregates(topLevelNumericAggInputs map[string]*gql.InputObject) []*gql.Field {
	topLevelSumField := gql.Field{
		Name:        request.SumFieldName,
		Description: schemaTypes.SumFieldDescription,
		Type:        gql.Float,
		Args:        gql.FieldConfigArgument{},
	}

	topLevelAverageField := gql.Field{
		Name:        request.AverageFieldName,
		Description: schemaTypes.AverageFieldDescription,
		Type:        gql.Float,
		Args:        gql.FieldConfigArgument{},
	}

	for name, inputObject := range topLevelNumericAggInputs {
		topLevelSumField.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
		topLevelAverageField.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
	}

	return []*gql.Field{&topLevelSumField, &topLevelAverageField}
}

func (g *Generator) genCountFieldConfig(obj *gql.Object) (gql.Field, error) {
	childTypesByFieldName := map[string]gql.Type{}

	for _, field := range obj.Fields() {
		// Only lists can be counted
		listType, isList := field.Type.(*gql.List)
		if !isList {
			continue
		}
		inputObjectName := genObjectCountName(listType.OfType.Name())
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
		Name:        request.CountFieldName,
		Description: schemaTypes.CountFieldDescription,
		Type:        gql.Int,
		Args:        gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
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
			inputObjectName = genNumericObjectSelectorName(listType.OfType.Name())
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
		Name:        request.SumFieldName,
		Description: schemaTypes.SumFieldDescription,
		Type:        gql.Float,
		Args:        gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
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
			inputObjectName = genNumericObjectSelectorName(listType.OfType.Name())
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
		Name:        request.AverageFieldName,
		Description: schemaTypes.AverageFieldDescription,
		Type:        gql.Float,
		Args:        gql.FieldConfigArgument{},
	}

	for name, inputObject := range childTypesByFieldName {
		field.Args[name] = schemaTypes.NewArgConfig(inputObject, inputObject.Description())
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
					request.LimitClause: &gql.InputObjectFieldConfig{
						Type:        gql.Int,
						Description: schemaTypes.LimitArgDescription,
					},
					request.OffsetClause: &gql.InputObjectFieldConfig{
						Type:        gql.Int,
						Description: schemaTypes.OffsetArgDescription,
					},
					request.OrderClause: &gql.InputObjectFieldConfig{
						Type:        g.manager.schema.TypeMap()["Ordering"],
						Description: schemaTypes.OrderArgDescription,
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
			request.LimitClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: schemaTypes.LimitArgDescription,
			},
			request.OffsetClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: schemaTypes.OffsetArgDescription,
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
				request.LimitClause: &gql.InputObjectFieldConfig{
					Type:        gql.Int,
					Description: schemaTypes.LimitArgDescription,
				},
				request.OffsetClause: &gql.InputObjectFieldConfig{
					Type:        gql.Int,
					Description: schemaTypes.OffsetArgDescription,
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
						fieldsEnumCfg.Values[request.CountFieldName] = &gql.EnumValueConfig{Value: request.CountFieldName}
					}
				}
			}
			// A child aggregate will always be aggregatable, as it can be present via an inner grouping
			fieldsEnumCfg.Values[request.SumFieldName] = &gql.EnumValueConfig{Value: request.SumFieldName}
			fieldsEnumCfg.Values[request.AverageFieldName] = &gql.EnumValueConfig{Value: request.AverageFieldName}

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
			request.LimitClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: schemaTypes.LimitArgDescription,
			},
			request.OffsetClause: &gql.InputObjectFieldConfig{
				Type:        gql.Int,
				Description: schemaTypes.OffsetArgDescription,
			},
			request.OrderClause: &gql.InputObjectFieldConfig{
				Type:        g.manager.schema.TypeMap()[genTypeName(obj, "OrderArg")],
				Description: schemaTypes.OrderArgDescription,
			},
		}, nil
	}

	return gql.NewInputObject(gql.InputObjectConfig{
		Name:   genNumericObjectSelectorName(obj.Name()),
		Fields: fieldThunk,
	})
}

func (g *Generator) appendCommitChildGroupField() {
	commitObject := g.manager.schema.TypeMap()[request.CommitTypeName]

	commitObject.(*gql.Object).Fields()[request.GroupFieldName] = &gql.FieldDefinition{
		Name:        request.GroupFieldName,
		Description: groupFieldDescription,
		Type:        gql.NewList(commitObject),
	}
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

	queryField := g.genTypeQueryableFieldList(obj, types)

	return queryField, nil
}

// GenerateMutationInputForGQLType creates all the mutation types and fields
// for the given graphQL object. It assumes that all the various
// filterArgs for the given type already exists, and will error otherwise.
func (g *Generator) GenerateMutationInputForGQLType(obj *gql.Object) ([]*gql.Field, error) {
	if obj.Error() != nil {
		return nil, obj.Error()
	}

	filterInputName := genTypeName(obj, filterInputNameSuffix)
	mutationInputName := genTypeName(obj, mutationInputNameSuffix)

	filterInput, ok := g.manager.schema.TypeMap()[filterInputName].(*gql.InputObject)
	if !ok {
		return nil, NewErrTypeNotFound(filterInputName)
	}

	mutationInput, ok := g.manager.schema.TypeMap()[mutationInputName]
	if !ok {
		return nil, NewErrTypeNotFound(mutationInputName)
	}

	explicitUserFieldsEnum := g.genUserExplicitTypeFieldsEnum(obj)

	g.manager.schema.TypeMap()[explicitUserFieldsEnum.Name()] = explicitUserFieldsEnum

	create := &gql.Field{
		Name:        "create_" + obj.Name(),
		Description: createDocumentDescription,
		Type:        gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			request.Input: schemaTypes.NewArgConfig(mutationInput, "Create a "+obj.Name()+" document"),
			request.Inputs: schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(mutationInput)),
				"Create "+obj.Name()+" documents"),
			request.EncryptDocArgName: schemaTypes.NewArgConfig(gql.Boolean, encryptArgDescription),
			request.EncryptFieldsArgName: schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(explicitUserFieldsEnum)),
				encryptFieldsArgDescription),
		},
	}

	update := &gql.Field{
		Name:        "update_" + obj.Name(),
		Description: updateDocumentsDescription,
		Type:        gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName:  schemaTypes.NewArgConfig(gql.ID, updateIDArgDescription),
			request.DocIDsArgName: schemaTypes.NewArgConfig(gql.NewList(gql.ID), updateIDsArgDescription),
			"filter":              schemaTypes.NewArgConfig(filterInput, updateFilterArgDescription),
			request.Input:         schemaTypes.NewArgConfig(mutationInput, "Update field values"),
		},
	}

	delete := &gql.Field{
		Name:        "delete_" + obj.Name(),
		Description: deleteDocumentsDescription,
		Type:        gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName:  schemaTypes.NewArgConfig(gql.ID, deleteIDArgDescription),
			request.DocIDsArgName: schemaTypes.NewArgConfig(gql.NewList(gql.ID), deleteIDsArgDescription),
			"filter":              schemaTypes.NewArgConfig(filterInput, deleteFilterArgDescription),
		},
	}

	return []*gql.Field{create, update, delete}, nil
}

func (g *Generator) genTypeFieldsEnum(obj *gql.Object) *gql.Enum {
	enumFieldsCfg := gql.EnumConfig{
		Name:   genTypeName(obj, typeFieldEnumSuffix),
		Values: gql.EnumValueConfigMap{},
	}

	for f, field := range obj.Fields() {
		enumFieldsCfg.Values[field.Name] = &gql.EnumValueConfig{Value: f}
	}

	return gql.NewEnum(enumFieldsCfg)
}

func (g *Generator) genUserExplicitTypeFieldsEnum(obj *gql.Object) *gql.Enum {
	enumFieldsCfg := gql.EnumConfig{
		Name:   genTypeName(obj, typeExplicitFieldEnumSuffix),
		Values: gql.EnumValueConfigMap{},
	}

	for f, field := range obj.Fields() {
		if strings.HasPrefix(field.Name, "_") {
			continue
		}
		enumFieldsCfg.Values[field.Name] = &gql.EnumValueConfig{Value: f}
	}

	return gql.NewEnum(enumFieldsCfg)
}

// input {Type.Name}FilterArg { ... }
func (g *Generator) genTypeFilterArgInput(obj *gql.Object) *gql.InputObject {
	var selfRefType *gql.InputObject

	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, filterInputNameSuffix),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(
		func() (gql.InputObjectConfigFieldMap, error) {
			fields := gql.InputObjectConfigFieldMap{}

			fields["_and"] = &gql.InputObjectFieldConfig{
				Description: schemaTypes.AndOperatorDescription,
				Type:        gql.NewList(selfRefType),
			}
			fields["_or"] = &gql.InputObjectFieldConfig{
				Description: schemaTypes.OrOperatorDescription,
				Type:        gql.NewList(selfRefType),
			}
			fields["_not"] = &gql.InputObjectFieldConfig{
				Description: schemaTypes.NotOperatorDescription,
				Type:        selfRefType,
			}

			// generate basic filter operator blocks
			// @todo: Extract object field loop into its own utility func
			for f, field := range obj.Fields() {
				if _, ok := request.ReservedFields[f]; ok && f != request.DocIDFieldName {
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
					fieldType := field.Type
					if l, isList := field.Type.(*gql.List); isList {
						// We want the FilterArg for the object, not the list of objects.
						fieldType = l.OfType
					}
					filterType, isFilterable := g.manager.schema.TypeMap()[genTypeName(fieldType, filterInputNameSuffix)]
					if !isFilterable {
						filterType = &gql.InputObjectField{}
					}
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: filterType,
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
		Name: fmt.Sprintf("%s%s", filterTypeName, filterInputNameSuffix),
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
			return nil, NewErrTypeNotFound(operatorBlockName)
		}

		operatorObject, isInputObj := operatorType.(*gql.InputObject)
		if !isInputObj {
			// This should be impossible
			return nil, client.NewErrUnexpectedType[*gql.InputObject]("operatorType", operatorType)
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
				if _, ok := request.ReservedFields[f]; ok && f != request.DocIDFieldName {
					continue
				}
				typeMap := g.manager.schema.TypeMap()
				configType, isOrderable := typeMap[genTypeName(field.Type, "OrderArg")]
				if gql.IsLeafType(field.Type) { // only Scalars, and enums
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: typeMap["Ordering"],
					}
				} else if isOrderable { // sub objects
					fields[field.Name] = &gql.InputObjectFieldConfig{
						Type: configType,
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
	obj *gql.Object,
	config queryInputTypeConfig,
) *gql.Field {
	name := obj.Name()

	// add the generated types to the type map
	g.manager.schema.TypeMap()[config.filter.Name()] = config.filter
	g.manager.schema.TypeMap()[config.groupBy.Name()] = config.groupBy
	g.manager.schema.TypeMap()[config.order.Name()] = config.order

	field := &gql.Field{
		Name:        name,
		Description: obj.Description(),
		Type:        gql.NewList(obj),
		Args: gql.FieldConfigArgument{
			request.DocIDArgName:  schemaTypes.NewArgConfig(gql.String, docIDArgDescription),
			request.DocIDsArgName: schemaTypes.NewArgConfig(gql.NewList(gql.NewNonNull(gql.String)), docIDsArgDescription),
			"cid":                 schemaTypes.NewArgConfig(gql.String, cidArgDescription),
			"filter":              schemaTypes.NewArgConfig(config.filter, selectFilterArgDescription),
			"groupBy": schemaTypes.NewArgConfig(
				gql.NewList(gql.NewNonNull(config.groupBy)),
				schemaTypes.GroupByArgDescription,
			),
			"order":              schemaTypes.NewArgConfig(config.order, schemaTypes.OrderArgDescription),
			request.ShowDeleted:  schemaTypes.NewArgConfig(gql.Boolean, showDeletedArgDescription),
			request.LimitClause:  schemaTypes.NewArgConfig(gql.Int, schemaTypes.LimitArgDescription),
			request.OffsetClause: schemaTypes.NewArgConfig(gql.Int, schemaTypes.OffsetArgDescription),
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

/* Example

typeDefs := ` ... `

ast, err := request.Parse(typeDefs)
types, err := buildTypesFromAST(ast)

types, err := GenerateDBQuerySchema(ast)
schemaManager.Update(types)

// request
q := query.Parse(qry)
qplan := planner.Plan(q, schemaManager.Schema)
resp := db.queryEngine.Execute(ctx, q, qplan)


*/
