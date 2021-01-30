package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/db/base"

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
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

	expandedTypes map[string]bool
}

// NewGenerator creates a new instance of the Generator
// from a given SchemaManager
func (m *SchemaManager) NewGenerator() *Generator {
	m.Generator = &Generator{
		manager:       m,
		expandedTypes: make(map[string]bool),
	}
	return m.Generator
}

// FromSDL generates the query type definitions from a
// encoded GraphQL Schema Definition Lanaguage string
func (g *Generator) FromSDL(schema string) ([]*gql.Object, error) {
	// parse to AST
	source := source.NewSource(&source.Source{
		Body: []byte(schema),
	})
	doc, err := parser.Parse(parser.ParseParams{
		Source: source,
	})
	if err != nil {
		return nil, err
	}
	// generate from AST
	return g.FromAST(doc)
}

// FromAST generates the query type definitions from a
// parsed GraphQL Schema Definition Language AST document
func (g *Generator) FromAST(document *ast.Document) ([]*gql.Object, error) {
	// build base types
	if err := g.buildTypesFromAST(document); err != nil {
		return nil, err
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// for each built type
	// 		generate query inputs
	queryType := g.manager.schema.QueryType()
	for _, t := range g.typeDefs {
		f, err := g.GenerateQueryInputForGQLType(t)
		if err != nil {
			return nil, err
		}
		queryType.AddFieldConfig(f.Name, f)
	}
	// resolve types
	if err := g.manager.ResolveTypes(); err != nil {
		return nil, err
	}

	// secondary pass to expand query collection type
	// argument inputs
	query := g.manager.schema.QueryType()
	collections := query.Fields()
	for _, def := range collections {
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

	return g.typeDefs, nil
}

func (g *Generator) expandInputArgument(obj *gql.Object) error {
	fields := obj.Fields()
	for f, def := range fields {
		switch t := def.Type.(type) {
		case *gql.Object:
			if _, complete := g.expandedTypes[obj.Name()]; complete {
				continue
			} else {
				g.expandedTypes[obj.Name()] = true
			}
			// make sure all the sub fields are expanded first
			if err := g.expandInputArgument(t); err != nil {
				return err
			}

			// new field object with arugments (single)
			expandedField, err := g.createExpandedFieldSingle(def, t)
			if err != nil {
				return err
			}

			// obj.AddFieldConfig(f, expandedField)
			// obj := g.manager.schema.Type(obj.Name()).(*gql.Object)
			obj.AddFieldConfig(f, expandedField)
			break
		case *gql.List: // new field object with aguments (list)
			listType := t.OfType
			if _, complete := g.expandedTypes[obj.Name()]; complete {
				continue
			} else {
				g.expandedTypes[obj.Name()] = true
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
			// todo: check if NonNull is possible here
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
func (g *Generator) buildTypesFromAST(document *ast.Document) error {
	// @todo: Check for duplicate named defined types in the TypeMap
	// get all the defined types from the AST
	for _, def := range document.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
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
			fieldsThunk := (gql.FieldsThunk)(func() gql.Fields {
				fields := gql.Fields{}

				// @todo: Check if this is a collection (relation) type
				// or just a embedded only type (which doesnt need a key)
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
						// @todo: Handle errors during type genation within a Thunk
						// panic(err)
					}

					// check if ttype is a Object value
					// if so, add appropriate relationship data
					// @todo check various directives for nature
					// of object relationship
					switch subobj := ttype.(type) {
					case *gql.Object:
						fields[fType.Name+"_id"] = &gql.Field{Type: gql.ID}

						// register the relation
						relName, err := genRelationName(objconf.Name, ttype.Name())
						if err != nil {
							// todo again handle errors
						}
						g.manager.Relations.RegisterSingle(relName, ttype.Name(), fType.Name, base.Meta_Relation_ONE)
					case *gql.List:
						ltype := subobj.OfType
						// register the relation
						relName, err := genRelationName(objconf.Name, ltype.Name())
						if err != nil {
							// todo again handle errors
						}
						g.manager.Relations.RegisterSingle(relName, ltype.Name(), fType.Name, base.Meta_Relation_MANY)
						break
					}

					fType.Type = ttype
					fields[fType.Name] = fType
				}

				return fields
			})

			objconf.Fields = fieldsThunk

			obj := gql.NewObject(objconf)
			g.manager.schema.TypeMap()[obj.Name()] = obj
			g.typeDefs = append(g.typeDefs, obj)
			// s.types = append(s.types, obj)
		}
	}

	return nil
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
		return nil, errors.New("No type found for given name")
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

	for i, field := range obj.Fields() {
		enumFieldsCfg.Values[field.Name] = &gql.EnumValueConfig{Value: i}
	}

	return gql.NewEnum(enumFieldsCfg)
}

// input {Type.Name}FilterArg { ... }
func (g *Generator) genTypeFilterArgInput(obj *gql.Object) *gql.InputObject {
	var selfRefType *gql.InputObject

	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterArg"),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(func() gql.InputObjectConfigFieldMap {
		fields := gql.InputObjectConfigFieldMap{}

		// @attention: do we need to explicity add our "sub types" to the TypeMap
		filterBaseArgType := g.genTypeFilterBaseArgInput(obj)
		g.manager.schema.AppendType(filterBaseArgType)

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
		for _, field := range obj.Fields() {

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
		return fields
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
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(func() gql.InputObjectConfigFieldMap {
		fields := gql.InputObjectConfigFieldMap{}

		for _, field := range obj.Fields() {
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

		return fields
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

// generate the type Query { ... }  field for the given type
func (g *Generator) genTypeQueryableField(obj *gql.Object, config queryInputTypeConfig) *gql.Field {
	name := strings.ToLower(obj.Name())

	// add the generated types to the type map
	// g.manager.schema.AppendType(config.filter)

	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: name,
		Type: obj,
		Args: gql.FieldConfigArgument{
			"filter": newArgConfig(config.filter),
		},
	}

	return field
}

func (g *Generator) genTypeQueryableFieldList(obj *gql.Object, config queryInputTypeConfig) *gql.Field {
	name := strings.ToLower(obj.Name())

	// add the generated types to the type map
	g.manager.schema.AppendType(config.filter)
	g.manager.schema.AppendType(config.groupBy)
	g.manager.schema.AppendType(config.having)
	g.manager.schema.AppendType(config.order)

	field := &gql.Field{
		// @todo: Handle collection name from @collection directive
		Name: name,
		Type: gql.NewList(obj),
		Args: gql.FieldConfigArgument{
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
	g.expandedTypes = make(map[string]bool)
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
