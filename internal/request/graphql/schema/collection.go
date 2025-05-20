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
	"fmt"
	"sort"
	"strconv"
	"strings"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
)

const (
	typeID       string = "ID"
	typeBoolean  string = "Boolean"
	typeInt      string = "Int"
	typeFloat    string = "Float"
	typeFloat32  string = "Float32"
	typeFloat64  string = "Float64"
	typeDateTime string = "DateTime"
	typeString   string = "String"
	typeBlob     string = "Blob"
	typeJSON     string = "JSON"
)

// this mapping is used to check that the default prop value
// matches the field type
var TypeToDefaultPropName = map[string]string{
	typeString:   types.DefaultDirectivePropString,
	typeBoolean:  types.DefaultDirectivePropBool,
	typeInt:      types.DefaultDirectivePropInt,
	typeFloat:    types.DefaultDirectivePropFloat,
	typeFloat32:  types.DefaultDirectivePropFloat32,
	typeFloat64:  types.DefaultDirectivePropFloat64,
	typeDateTime: types.DefaultDirectivePropDateTime,
	typeJSON:     types.DefaultDirectivePropJSON,
	typeBlob:     types.DefaultDirectivePropBlob,
}

type typeDefinition struct {
	Name        *ast.Name
	Description *ast.StringValue
	Directives  []*ast.Directive
	Fields      []*ast.FieldDefinition
	IsInterface bool
}

func newInterfaceDefinition(def *ast.InterfaceDefinition) *typeDefinition {
	return &typeDefinition{
		Name:        def.Name,
		Description: def.Description,
		Directives:  def.Directives,
		Fields:      def.Fields,
		IsInterface: true,
	}
}

func newObjectDefinition(def *ast.ObjectDefinition) *typeDefinition {
	return &typeDefinition{
		Name:        def.Name,
		Description: def.Description,
		Directives:  def.Directives,
		Fields:      def.Fields,
	}
}

// fromAst parses a GQL AST into a set of collection versions.
func fromAst(doc *ast.Document) (
	[]core.Collection,
	error,
) {
	results := []core.Collection{}
	cTypeByFieldNameByObjName := map[string]map[string]client.CType{}

	for _, def := range doc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			td := newObjectDefinition(defType)
			result, err := fromAstDefinition(td, cTypeByFieldNameByObjName)
			if err != nil {
				return nil, err
			}

			results = append(results, result)

		case *ast.InterfaceDefinition:
			td := newInterfaceDefinition(defType)
			result, err := fromAstDefinition(td, cTypeByFieldNameByObjName)
			if err != nil {
				return nil, err
			}

			results = append(results, result)

		default:
			// Do nothing, ignore it and continue
			continue
		}
	}

	// The details on the relations between objects depend on both sides
	// of the relationship.  The relation manager handles this, and must be applied
	// after all the collections have been processed.
	err := finalizeRelations(results, cTypeByFieldNameByObjName)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// fromAstDefinition parses a AST object definition into a set of collection versions.
func fromAstDefinition(
	def *typeDefinition,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) (core.Collection, error) {
	schemaFieldDescriptions := []client.SchemaFieldDescription{
		{
			Name: request.DocIDFieldName,
			Kind: client.FieldKind_DocID,
			Typ:  client.NONE_CRDT,
		},
	}
	collectionFieldDescriptions := []client.CollectionFieldDescription{
		{
			Name: request.DocIDFieldName,
		},
	}

	policyDescription := immutable.None[client.PolicyDescription]()

	indexes := []client.IndexCreateRequest{}
	vectorEmbeddings := []client.VectorEmbeddingDescription{}
	encryptedIndexes := []client.EncryptedIndexCreateRequest{}
	for _, field := range def.Fields {
		tmpSchemaFieldDescriptions, tmpCollectionFieldDescriptions, err := fieldsFromAST(
			field,
			def.Name.Value,
			cTypeByFieldNameByObjName,
		)
		if err != nil {
			return core.Collection{}, err
		}

		schemaFieldDescriptions = append(schemaFieldDescriptions, tmpSchemaFieldDescriptions...)
		collectionFieldDescriptions = append(collectionFieldDescriptions, tmpCollectionFieldDescriptions...)

		for _, directive := range field.Directives {
			switch directive.Name.Value {
			case types.IndexDirectiveLabel:
				index, err := indexFromAST(directive, field)
				if err != nil {
					return core.Collection{}, err
				}
				indexes = append(indexes, index)
			case types.VectorEmbeddingDirectiveLabel:
				embedding, err := vectorEmbeddingFromAST(directive, field)
				if err != nil {
					return core.Collection{}, err
				}
				vectorEmbeddings = append(vectorEmbeddings, embedding)
			case types.EncryptedIndexDirectiveLabel:
				encryptedIndex, err := encryptedIndexFromAST(directive, field)
				if err != nil {
					return core.Collection{}, err
				}
				encryptedIndexes = append(encryptedIndexes, encryptedIndex)
			}
		}
	}

	// sort the fields lexicographically
	sort.Slice(schemaFieldDescriptions, func(i, j int) bool {
		// make sure that the _docID is always at the beginning
		if schemaFieldDescriptions[i].Name == request.DocIDFieldName {
			return true
		} else if schemaFieldDescriptions[j].Name == request.DocIDFieldName {
			return false
		}
		return schemaFieldDescriptions[i].Name < schemaFieldDescriptions[j].Name
	})
	sort.Slice(collectionFieldDescriptions, func(i, j int) bool {
		// make sure that the _docID is always at the beginning
		if collectionFieldDescriptions[i].Name == request.DocIDFieldName {
			return true
		} else if collectionFieldDescriptions[j].Name == request.DocIDFieldName {
			return false
		}
		return collectionFieldDescriptions[i].Name < collectionFieldDescriptions[j].Name
	})

	isMaterialized := immutable.None[bool]()
	var isBranchable bool
	for _, directive := range def.Directives {
		switch directive.Name.Value {
		case types.IndexDirectiveLabel:
			index, err := indexFromAST(directive, nil)
			if err != nil {
				return core.Collection{}, err
			}
			indexes = append(indexes, index)

		case types.PolicySchemaDirectiveLabel:
			policy, err := policyFromAST(directive)
			if err != nil {
				return core.Collection{}, err
			}
			policyDescription = immutable.Some(policy)

		case types.MaterializedDirectiveLabel:
			if isMaterialized.Value() {
				continue
			}

			explicitIsMaterialized := immutable.None[bool]()
			for _, arg := range directive.Arguments {
				if arg.Name.Value == types.MaterializedDirectivePropIf {
					explicitIsMaterialized = immutable.Some(arg.Value.GetValue().(bool))
					break
				}
			}

			if explicitIsMaterialized.HasValue() {
				isMaterialized = immutable.Some(isMaterialized.Value() || explicitIsMaterialized.Value())
			} else {
				isMaterialized = immutable.Some(true)
			}

		case types.BranchableDirectiveLabel:
			if isBranchable {
				continue
			}

			explicitIsBranchable := immutable.None[bool]()

			for _, arg := range directive.Arguments {
				if arg.Name.Value == types.BranchableDirectivePropIf {
					explicitIsBranchable = immutable.Some(arg.Value.GetValue().(bool))
					break
				}
			}

			isBranchable = !explicitIsBranchable.HasValue() || explicitIsBranchable.Value()
		}
	}

	return core.Collection{
		Definition: client.CollectionDefinition{
			Version: client.CollectionVersion{
				Name:             def.Name.Value,
				Policy:           policyDescription,
				Fields:           collectionFieldDescriptions,
				IsMaterialized:   !isMaterialized.HasValue() || isMaterialized.Value(),
				IsBranchable:     isBranchable,
				IsEmbeddedOnly:   def.IsInterface,
				IsActive:         true,
				VectorEmbeddings: vectorEmbeddings,
			},
			Schema: client.SchemaDescription{
				Name:   def.Name.Value,
				Fields: schemaFieldDescriptions,
			},
		},
		CreateIndexes:          indexes,
		CreateEncryptedIndexes: encryptedIndexes,
	}, nil
}

// IsValidIndexName returns true if the name is a valid index name.
// Valid index names must start with a letter or underscore, and can
// contain letters, numbers, and underscores.
func IsValidIndexName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if name[0] != '_' && (name[0] < 'a' || name[0] > 'z') && (name[0] < 'A' || name[0] > 'Z') {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}

func indexFromAST(directive *ast.Directive, fieldDef *ast.FieldDefinition) (client.IndexCreateRequest, error) {
	var name string
	var unique bool

	var direction *ast.EnumValue
	var includes *ast.ListValue

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			nameVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.IndexCreateRequest{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value
			if !IsValidIndexName(name) {
				return client.IndexCreateRequest{}, NewErrIndexWithInvalidName(name)
			}

		case types.IndexDirectivePropIncludes:
			includesVal, ok := arg.Value.(*ast.ListValue)
			if !ok {
				return client.IndexCreateRequest{}, ErrIndexWithInvalidArg
			}
			includes = includesVal

		case types.IndexDirectivePropDirection:
			directionVal, ok := arg.Value.(*ast.EnumValue)
			if !ok {
				return client.IndexCreateRequest{}, ErrIndexWithInvalidArg
			}
			direction = directionVal

		case types.IndexDirectivePropUnique:
			uniqueVal, ok := arg.Value.(*ast.BooleanValue)
			if !ok {
				return client.IndexCreateRequest{}, ErrIndexWithInvalidArg
			}
			unique = uniqueVal.Value

		default:
			return client.IndexCreateRequest{}, ErrIndexWithUnknownArg
		}
	}

	var containsField bool
	var fields []client.IndexedFieldDescription

	if includes != nil {
		for _, include := range includes.Values {
			field, err := indexFieldFromAST(include, direction)
			if err != nil {
				return client.IndexCreateRequest{}, err
			}
			if fieldDef != nil && fieldDef.Name.Value == field.Name {
				containsField = true
			}
			fields = append(fields, field)
		}
	}

	// if the directive is applied to a field and
	// the field is not in the includes list
	// implicitly add it as the first entry
	if !containsField && fieldDef != nil {
		field := client.IndexedFieldDescription{
			Name: fieldDef.Name.Value,
		}
		if direction != nil {
			field.Descending = direction.Value == types.FieldOrderDESC
		}
		fields = append([]client.IndexedFieldDescription{field}, fields...)
	}

	if len(fields) == 0 {
		return client.IndexCreateRequest{}, ErrIndexMissingFields
	}

	return client.IndexCreateRequest{
		Name:   name,
		Fields: fields,
		Unique: unique,
	}, nil
}

func indexFieldFromAST(value ast.Value, defaultDirection *ast.EnumValue) (client.IndexedFieldDescription, error) {
	argTypeObject, ok := value.(*ast.ObjectValue)
	if !ok {
		return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
	}

	var name string
	var direction *ast.EnumValue

	for _, field := range argTypeObject.Fields {
		switch field.Name.Value {
		case types.IncludesPropField:
			nameVal, ok := field.Value.(*ast.StringValue)
			if !ok {
				return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value

		case types.IncludesPropDirection:
			directionVal, ok := field.Value.(*ast.EnumValue)
			if !ok {
				return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
			}
			direction = directionVal

		default:
			return client.IndexedFieldDescription{}, ErrIndexWithUnknownArg
		}
	}

	var descending bool
	// if the direction is explicitly set use that value, otherwise
	// if the default direction was set on the index use that value
	if direction != nil {
		descending = direction.Value == types.FieldOrderDESC
	} else if defaultDirection != nil {
		descending = defaultDirection.Value == types.FieldOrderDESC
	}

	return client.IndexedFieldDescription{
		Name:       name,
		Descending: descending,
	}, nil
}

func defaultFromAST(
	field *ast.FieldDefinition,
	directive *ast.Directive,
) (any, error) {
	astNamed, ok := field.Type.(*ast.Named)
	if !ok {
		return nil, NewErrDefaultValueNotAllowed(field.Name.Value, field.Type.String())
	}
	propName, ok := TypeToDefaultPropName[astNamed.Name.Value]
	if !ok {
		return nil, NewErrDefaultValueNotAllowed(field.Name.Value, astNamed.Name.Value)
	}
	if len(directive.Arguments) != 1 {
		return nil, NewErrDefaultValueOneArg(field.Name.Value)
	}
	arg := directive.Arguments[0]
	if propName != arg.Name.Value {
		return nil, NewErrDefaultValueType(field.Name.Value, propName, arg.Name.Value)
	}
	var value any
	switch propName {
	case types.DefaultDirectivePropInt:
		value = gql.Int.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropFloat:
		value = gql.Float.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropFloat32:
		value = types.Float32.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropFloat64:
		value = types.Float64.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropBool:
		value = gql.Boolean.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropString:
		value = gql.String.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropDateTime:
		value = gql.DateTime.ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropJSON:
		value = types.JSONScalarType().ParseLiteral(arg.Value, nil)
	case types.DefaultDirectivePropBlob:
		value = types.BlobScalarType().ParseLiteral(arg.Value, nil)
	}
	// If the value is nil, then parsing has failed, or a nil value was provided.
	// Since setting a default value to nil is the same as not providing one,
	// it is safer to return an error to let the user know something is wrong.
	if value == nil {
		return nil, NewErrDefaultValueInvalid(field.Name.Value, propName)
	}
	return value, nil
}

func encryptedIndexFromAST(
	directive *ast.Directive,
	fieldDef *ast.FieldDefinition,
) (client.EncryptedIndexCreateRequest, error) {
	encryptedIndex := client.EncryptedIndexCreateRequest{
		FieldName: fieldDef.Name.Value,
	}

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.EncryptedIndexDirectivePropType:
			typeVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.EncryptedIndexCreateRequest{}, NewErrEncryptedIndexWithInvalidArg(fieldDef.Name.Value)
			}

			// Currently only equality is supported
			if typeVal.Value != string(client.EncryptedIndexTypeEquality) {
				return client.EncryptedIndexCreateRequest{}, NewErrEncryptedIndexTypeNotSupported(typeVal.Value)
			}
			encryptedIndex.Type = client.EncryptedIndexType(typeVal.Value)

		default:
			return client.EncryptedIndexCreateRequest{}, NewErrEncryptedIndexWithUnknownArg(arg.Name.Value)
		}
	}

	return encryptedIndex, nil
}

func fieldsFromAST(
	field *ast.FieldDefinition,
	hostObjectName string,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) ([]client.SchemaFieldDescription, []client.CollectionFieldDescription, error) {
	kind, err := astTypeToKind(hostObjectName, field)
	if err != nil {
		return nil, nil, err
	}

	cType, err := setCRDTType(field, kind)
	if err != nil {
		return nil, nil, err
	}

	hostMap := cTypeByFieldNameByObjName[hostObjectName]
	if hostMap == nil {
		hostMap = map[string]client.CType{}
		cTypeByFieldNameByObjName[hostObjectName] = hostMap
	}
	hostMap[field.Name.Value] = cType

	var defaultValue any
	var constraints constraintDescription
	for _, directive := range field.Directives {
		switch directive.Name.Value {
		case types.DefaultDirectiveLabel:
			defaultValue, err = defaultFromAST(field, directive)
			if err != nil {
				return nil, nil, err
			}
		case types.ConstraintsDirectiveLabel:
			constraints, err = contraintsFromAST(kind, directive)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	schemaFieldDescriptions := []client.SchemaFieldDescription{}
	collectionFieldDescriptions := []client.CollectionFieldDescription{}

	if namedKind, ok := kind.(*client.NamedKind); ok {
		relationName, err := getRelationshipName(field, hostObjectName, namedKind.Name)
		if err != nil {
			return nil, nil, err
		}

		if kind.IsArray() {
			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         field.Name.Value,
					Kind:         immutable.Some(kind),
					RelationName: immutable.Some(relationName),
				},
			)
		} else {
			idFieldName := fmt.Sprintf("%s_id", field.Name.Value)

			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         idFieldName,
					Kind:         immutable.Some[client.FieldKind](client.FieldKind_DocID),
					RelationName: immutable.Some(relationName),
				},
			)

			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         field.Name.Value,
					Kind:         immutable.Some(kind),
					RelationName: immutable.Some(relationName),
				},
			)

			if _, exists := findDirective(field, "primary"); exists {
				// Only primary fields exist on the schema.  If primary is automatically set
				// (e.g. for one-many) a later step will add this property.
				schemaFieldDescriptions = append(
					schemaFieldDescriptions,
					client.SchemaFieldDescription{
						Name: field.Name.Value,
						Kind: kind,
						Typ:  cType,
					},
				)
			}
		}
	} else {
		schemaFieldDescriptions = append(
			schemaFieldDescriptions,
			client.SchemaFieldDescription{
				Name: field.Name.Value,
				Kind: kind,
				Typ:  cType,
			},
		)

		collectionFieldDescriptions = append(
			collectionFieldDescriptions,
			client.CollectionFieldDescription{
				Name:         field.Name.Value,
				DefaultValue: defaultValue,
				Size:         constraints.Size,
			},
		)
	}

	return schemaFieldDescriptions, collectionFieldDescriptions, nil
}

// policyFromAST returns the policy description after parsing but the validation
// is not done yet on the values that are returned. This is because we need acp to do that.
func policyFromAST(directive *ast.Directive) (client.PolicyDescription, error) {
	policyDesc := client.PolicyDescription{}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.PolicySchemaDirectivePropID:
			policyIDProp, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.PolicyDescription{}, ErrPolicyInvalidIDProp
			}
			policyDesc.ID = policyIDProp.Value
		case types.PolicySchemaDirectivePropResource:
			policyResourceProp, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.PolicyDescription{}, ErrPolicyInvalidResourceProp
			}
			policyDesc.ResourceName = policyResourceProp.Value
		default:
			return client.PolicyDescription{}, ErrPolicyWithUnknownArg
		}
	}
	return policyDesc, nil
}

func vectorEmbeddingFromAST(
	directive *ast.Directive,
	fieldDef *ast.FieldDefinition,
) (client.VectorEmbeddingDescription, error) {
	embedding := client.VectorEmbeddingDescription{
		FieldName: fieldDef.Name.Value,
	}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.VectorEmbeddingDirectivePropFields:
			val := arg.Value.(*ast.ListValue)
			fields := make([]string, len(val.Values))
			for i, untypedField := range val.Values {
				fields[i] = untypedField.(*ast.StringValue).Value
			}
			embedding.Fields = fields
		case types.VectorEmbeddingDirectivePropModel:
			embedding.Model = arg.Value.(*ast.StringValue).Value
		case types.VectorEmbeddingDirectivePropProvider:
			embedding.Provider = arg.Value.(*ast.StringValue).Value
		case types.VectorEmbeddingDirectivePropTemplate:
			embedding.Template = arg.Value.(*ast.StringValue).Value
		case types.VectorEmbeddingDirectivePropURL:
			embedding.URL = arg.Value.(*ast.StringValue).Value
		}
	}
	return embedding, nil
}

type constraintDescription struct {
	Size int
}

func contraintsFromAST(kind client.FieldKind, directive *ast.Directive) (constraintDescription, error) {
	constraints := constraintDescription{}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.ConstraintsDirectivePropSize:
			if !kind.IsArray() {
				return constraintDescription{}, NewErrInvalidTypeForContraint(kind)
			}
			size, err := strconv.Atoi(arg.Value.(*ast.IntValue).Value)
			if err != nil {
				return constraintDescription{}, err
			}
			constraints.Size = size
		}
	}
	return constraints, nil
}

func setCRDTType(field *ast.FieldDefinition, kind client.FieldKind) (client.CType, error) {
	if directive, exists := findDirective(field, "crdt"); exists {
		for _, arg := range directive.Arguments {
			switch arg.Name.Value {
			case "type":
				cTypeString := arg.Value.GetValue().(string)
				cType, validCRDTEnum := types.CRDTEnum().ParseValue(cTypeString).(client.CType)
				if !validCRDTEnum {
					return 0, client.NewErrInvalidCRDTType(field.Name.Value, cTypeString)
				}
				if !cType.IsCompatibleWith(kind) {
					return 0, client.NewErrCRDTKindMismatch(cType.String(), kind.String())
				}
				return cType, nil
			}
		}
	}

	if kind.IsObject() {
		if kind.IsArray() {
			return client.NONE_CRDT, nil
		}
		return client.LWW_REGISTER, nil
	}

	return defaultCRDTForFieldKind[kind], nil
}

func astTypeToKind(
	hostObjectName string,
	field *ast.FieldDefinition,
) (client.FieldKind, error) {
	switch astTypeVal := field.Type.(type) {
	case *ast.List:
		switch innerAstTypeVal := astTypeVal.Type.(type) {
		case *ast.NonNull:
			switch innerAstTypeVal.Type.(*ast.Named).Name.Value {
			case typeBoolean:
				return client.FieldKind_BOOL_ARRAY, nil
			case typeInt:
				return client.FieldKind_INT_ARRAY, nil
			case typeFloat, typeFloat64:
				return client.FieldKind_FLOAT64_ARRAY, nil
			case typeFloat32:
				return client.FieldKind_FLOAT32_ARRAY, nil
			case typeString:
				return client.FieldKind_STRING_ARRAY, nil
			default:
				return client.FieldKind_None, NewErrNonNullForTypeNotSupported(innerAstTypeVal.Type.(*ast.Named).Name.Value)
			}

		default:
			switch astTypeVal.Type.(*ast.Named).Name.Value {
			case typeBoolean:
				return client.FieldKind_NILLABLE_BOOL_ARRAY, nil
			case typeInt:
				return client.FieldKind_NILLABLE_INT_ARRAY, nil
			case typeFloat, typeFloat64:
				return client.FieldKind_NILLABLE_FLOAT64_ARRAY, nil
			case typeFloat32:
				return client.FieldKind_NILLABLE_FLOAT32_ARRAY, nil
			case typeString:
				return client.FieldKind_NILLABLE_STRING_ARRAY, nil
			default:
				return client.NewNamedKind(astTypeVal.Type.(*ast.Named).Name.Value, true), nil
			}
		}

	case *ast.Named:
		switch astTypeVal.Name.Value {
		case typeID:
			return client.FieldKind_DocID, nil
		case typeBoolean:
			return client.FieldKind_NILLABLE_BOOL, nil
		case typeInt:
			return client.FieldKind_NILLABLE_INT, nil
		case typeFloat, typeFloat64:
			return client.FieldKind_NILLABLE_FLOAT64, nil
		case typeFloat32:
			return client.FieldKind_NILLABLE_FLOAT32, nil
		case typeDateTime:
			return client.FieldKind_NILLABLE_DATETIME, nil
		case typeString:
			return client.FieldKind_NILLABLE_STRING, nil
		case typeBlob:
			return client.FieldKind_NILLABLE_BLOB, nil
		case typeJSON:
			return client.FieldKind_NILLABLE_JSON, nil
		default:
			return client.NewNamedKind(astTypeVal.Name.Value, false), nil
		}

	case *ast.NonNull:
		return client.FieldKind_None, ErrNonNullNotSupported

	default:
		if field.Type == nil {
			return client.FieldKind_None, NewErrFieldTypeNotSpecified(hostObjectName, field.Name.Value)
		}
		return client.FieldKind_None, NewErrTypeNotFound(field.Type.String())
	}
}

func findDirective(field *ast.FieldDefinition, directiveName string) (*ast.Directive, bool) {
	for _, directive := range field.Directives {
		if directive.Name.Value == directiveName {
			return directive, true
		}
	}
	return nil, false
}

// Gets the name of the relationship. Will return the provided name if one is specified,
// otherwise will generate one
func getRelationshipName(
	field *ast.FieldDefinition,
	hostName string,
	targetName string,
) (string, error) {
	// search for a @relation directive name, and return it if found
	for _, directive := range field.Directives {
		if directive.Name.Value == "relation" {
			for _, argument := range directive.Arguments {
				if argument.Name.Value == "name" {
					name, isString := argument.Value.GetValue().(string)
					if !isString {
						return "", client.NewErrUnexpectedType[string]("Relationship name", argument.Value.GetValue())
					}
					return name, nil
				}
			}
		}
	}

	// if no name is provided, generate one
	return genRelationName(hostName, targetName)
}

func genRelationName(t1, t2 string) (string, error) {
	if t1 == "" || t2 == "" {
		return "", client.NewErrUninitializeProperty("genRelationName", "relation types")
	}
	t1 = strings.ToLower(t1)
	t2 = strings.ToLower(t2)

	if i := strings.Compare(t1, t2); i < 0 {
		return fmt.Sprintf("%s_%s", t1, t2), nil
	}
	return fmt.Sprintf("%s_%s", t2, t1), nil
}

func finalizeRelations(
	results []core.Collection,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) error {
	for i, result := range results {
		if result.Definition.Version.IsEmbeddedOnly {
			// Embedded objects are simpler and require no addition work
			continue
		}

		for _, field := range result.Definition.Version.Fields {
			if !field.Kind.HasValue() {
				continue
			}

			namedKind, ok := field.Kind.Value().(*client.NamedKind)
			if !ok || namedKind.IsArray() {
				// We only need to process the primary side of a relation here, if the field is not a relation
				// or if it is an array, we can skip it.
				continue
			}

			var otherColDefinition immutable.Option[client.CollectionDefinition]
			for _, otherDef := range results {
				// Check the 'other' schema name, there can only be a one-one mapping in an SDL.
				if otherDef.Definition.Version.Name == namedKind.Name {
					otherColDefinition = immutable.Some(otherDef.Definition)
					break
				}
			}

			if !otherColDefinition.HasValue() {
				// If the other collection is not found here we skip this field.  Whilst this almost certainly means the SDL
				// is invalid, validating anything beyond SDL syntax is not the responsibility of this package.
				continue
			}

			otherColFieldDescription, hasOtherColFieldDescription := otherColDefinition.Value().Version.GetFieldByRelation(
				field.RelationName.Value(),
				result.Definition.GetName(),
				field.Name,
			)

			if !hasOtherColFieldDescription || otherColFieldDescription.Kind.Value().IsArray() {
				if _, exists := result.Definition.Schema.GetFieldByName(field.Name); !exists {
					// Relations only defined on one side of the object are possible, and so if this is one of them
					// or if the other side is an array, we need to add the field to the schema (is primary side)
					// if the field has not been explicitly declared by the user.
					result.Definition.Schema.Fields = append(
						result.Definition.Schema.Fields,
						client.SchemaFieldDescription{
							Name: field.Name,
							Kind: field.Kind.Value(),
							Typ:  cTypeByFieldNameByObjName[result.Definition.Version.Name][field.Name],
						},
					)
				}
			}

			if !otherColDefinition.Value().Version.IsEmbeddedOnly {
				var schemaFieldIndex int
				var schemaFieldExists bool
				for i, schemaField := range result.Definition.Schema.Fields {
					if schemaField.Name == field.Name {
						schemaFieldIndex = i
						schemaFieldExists = true
						break
					}
				}

				if schemaFieldExists {
					idFieldName := fmt.Sprintf("%s_id", field.Name)

					if _, idFieldExists := result.Definition.Schema.GetFieldByName(idFieldName); !idFieldExists {
						existingFields := result.Definition.Schema.Fields
						result.Definition.Schema.Fields = make([]client.SchemaFieldDescription, len(result.Definition.Schema.Fields)+1)
						copy(result.Definition.Schema.Fields, existingFields[:schemaFieldIndex+1])
						copy(result.Definition.Schema.Fields[schemaFieldIndex+2:], existingFields[schemaFieldIndex+1:])

						// An _id field is added for every 1-1 or 1-N relationship from this object if the relation
						// does not point to an embedded object.
						//
						// It is inserted immediately after the object field to make things nicer for the user.
						result.Definition.Schema.Fields[schemaFieldIndex+1] = client.SchemaFieldDescription{
							Name: idFieldName,
							Kind: client.FieldKind_DocID,
							Typ:  defaultCRDTForFieldKind[client.FieldKind_DocID],
						}
					}
				}
			}

			results[i] = result
		}
	}

	return nil
}
