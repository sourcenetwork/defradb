package graphql

import (
	"errors"
	"fmt"

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

// @todo: Add Schema Directives (IE: relation, etc..)

// Given a parsed AST of a developer defined types
// extract and return the correct objectType(s)
func (s *SchemaManager) buildTypesFromAST(document *ast.Document) error {
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
				ttype, err := astNodeToGqlType(s.typeMap, t)
				if err != nil {
					return err
				}
				fType.Type = ttype
			}

			objconf.Fields = fields

			obj := gql.NewObject(objconf)
			s.typeMap[obj.Name()] = obj
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

// GenerateSchemaForGQLType is the main generation function
// for creating the full DefraDB Query schema for a given
// developer defined type
func (s *SchemaManager) GenerateSchemaForGQLType(obj *gql.Object) {}

// enum {Type.Name}Fields { ... }
func (s *SchemaManager) genTypeFieldsEnum(obj gql.Object) *gql.Enum {
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
func (s *SchemaManager) genTypeFilterArgInput(obj gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterArg"),
	}
	fields := gql.InputObjectConfigFieldMap{}

	filterBaseArgType := s.genTypeFilterBaseArgInput(obj)

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

	// @todo: Handle Sub Object relational types using type indirection
	// @body: Currently, we can't support embedded types or relations in
	// our generation code because the GraphQL lib requires defined types
	// to exist before referencing them. However, if we define type A,
	// with another type embedded type B inside of it, when we generate the
	// schema types for A, we don't have the necessary matching types
	// for B yet. Need to implement some kind of type indirection that can
	// be effectiently resolved at a later time.

	// generate basic filter operator blocks for all the Leaf types
	// (scalars + enums)
	// @todo: Extract object field loop into its own utility func
	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: s.typeMap[field.Type.Name()+"OperatorBlock"],
			}
		}
	}

	// add objects (relations)

	return gql.NewInputObject(inputCfg)
}

// input {Type.Name}FilterBaseArg { ... }
func (s *SchemaManager) genTypeFilterBaseArgInput(obj gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "FilterBaseArg"),
	}
	fields := gql.InputObjectConfigFieldMap{}
	// generate basic filter operator blocks for all the Leaf types
	// (scalars + enums)
	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: s.typeMap[field.Type.Name()+"OperatorBlock"],
			}
		}
	}

	return gql.NewInputObject(inputCfg)
}

// query spec - sec N
func (s *SchemaManager) genTypeHavingArgInput(obj gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "HavingArg"),
	}
	fields := gql.InputObjectConfigFieldMap{}
	havingBlock := s.genTypeHavingBlockInput(obj)

	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: havingBlock,
			}
		}
	}

	return gql.NewInputObject(inputCfg)
}

func (s *SchemaManager) genTypeHavingBlockInput(obj gql.Object) *gql.InputObject {
	inputCfg := gql.InputObjectConfig{
		Name: genTypeName(obj, "HavingBlock"),
	}
	fields := gql.InputObjectConfigFieldMap{}

	for _, field := range obj.Fields() {
		if gql.IsLeafType(field.Type) { // only Scalars, and enums
			fields[field.Name] = &gql.InputObjectFieldConfig{
				Type: s.typeMap["FloatOperatorBlock"],
			}
		}
	}

	return gql.NewInputObject(inputCfg)
}

func (s *SchemaManager) genTypeOrderArgInput(obj gql.Object) *gql.InputObject {
	return nil
}

func (s *SchemaManager) genTypeQueryCollectionField(obj gql.Object) *gql.Object {
	return nil
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
