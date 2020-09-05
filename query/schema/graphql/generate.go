package graphql

import (
	"errors"
	"fmt"
	"strings"

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
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
}

// NewGenerator creates a new instance of the Generator
// from a given SchemaManager
func NewGenerator(manager *SchemaManager) *Generator {
	return &Generator{
		manager: manager,
	}
}

// FromString generates the query type definitions from a
// encoded GraphQL Schema Definition Lanaguage string
func (g *Generator) FromString(source string) error {
	return nil
}

// FromAST generates the query type definitions from a
// parsed GraphQL Schema Definition Language AST document
func (g *Generator) FromAST(document *ast.Document) error {
	return nil
}

// @todo: Add Schema Directives (IE: relation, etc..)

// @todo: Add validation support for the AST
// @body: Type generation is only supported for Object type definitions.
// Unions, Interfaces, etc are not currently supported.

// Given a parsed AST of a developer defined types
// extract and return the correct objectType(s)
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

			fields := gql.Fields{}
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
					return err
				}
				fType.Type = ttype
			}

			objconf.Fields = fields

			obj := gql.NewObject(objconf)
			g.manager.schema.TypeMap()[obj.Name()] = obj
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

// GenerateSchemaForGQLType is the main generation function
// for creating the full DefraDB Query schema for a given
// developer defined type
func (g *Generator) GenerateSchemaForGQLType(obj *gql.Object) {
	types := queryInputTypeConfig{}
	types.filter = g.genTypeFilterArgInput(*obj)
	types.groupBy = g.genTypeFieldsEnum(*obj)
	types.having = g.genTypeHavingArgInput(*obj)
	types.order = g.genTypeOrderArgInput(*obj)

	// @attention: return something lol
}

// enum {Type.Name}Fields { ... }
func (g *Generator) genTypeFieldsEnum(obj gql.Object) *gql.Enum {
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
func (g *Generator) genTypeFilterArgInput(obj gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterArg"),
	}
	fieldThunk := (gql.InputObjectConfigFieldMapThunk)(func() gql.InputObjectConfigFieldMap {
		fields := gql.InputObjectConfigFieldMap{}

		// @attention: do we need to explicity add our "sub types" to the TypeMap
		filterBaseArgType := g.genTypeFilterBaseArgInput(obj)
		g.manager.schema.TypeMap()[filterBaseArgType.Name()] = filterBaseArgType

		// conditionals
		fields["_and"] = &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(filterBaseArgType)),
		}
		fields["_or"] = &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(filterBaseArgType)),
		}
		fields["_not"] = &gql.InputObjectFieldConfig{
			Type: filterBaseArgType,
		}

		// generate basic filter operator blocks for all the Leaf types
		// (scalars + enums)
		// @todo: Extract object field loop into its own utility func
		for _, field := range obj.Fields() {
			if gql.IsLeafType(field.Type) { // only Scalars, and enums
				fields[field.Name] = &gql.InputObjectFieldConfig{
					Type: g.manager.schema.TypeMap()[field.Type.Name()+"OperatorBlock"],
				}
			}
		}

		// add objects (relations)

		return fields
	})

	// add the fields thunker
	inputCfg.Fields = fieldThunk
	return gql.NewInputObject(inputCfg)
}

// input {Type.Name}FilterBaseArg { ... }
func (g *Generator) genTypeFilterBaseArgInput(obj gql.Object) *gql.InputObject {
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
func (g *Generator) genTypeHavingArgInput(obj gql.Object) *gql.InputObject {
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

func (g *Generator) genTypeHavingBlockInput(obj gql.Object) *gql.InputObject {
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

func (g *Generator) genTypeOrderArgInput(obj gql.Object) *gql.InputObject {
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
			}
		}

		// add sub objects

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
func (g *Generator) genTypeQueryCollectionField(obj *gql.Object, config queryInputTypeConfig) *gql.Field {
	collectionName := strings.ToLower(obj.Name())
	fieldConfig := &gql.Field{
		Name: collectionName,
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

	return fieldConfig
}

func newArgConfig(t gql.Input) *gql.ArgumentConfig {
	return &gql.ArgumentConfig{
		Type: t,
	}
}

func genTypeName(obj gql.Object, name string) string {
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
