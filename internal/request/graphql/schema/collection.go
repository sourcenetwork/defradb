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
	"sort"
	"strings"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
)

const (
	typeID       string = "ID"
	typeBoolean  string = "Boolean"
	typeInt      string = "Int"
	typeFloat    string = "Float"
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
	typeDateTime: types.DefaultDirectivePropDateTime,
	typeJSON:     types.DefaultDirectivePropJSON,
	typeBlob:     types.DefaultDirectivePropBlob,
}

// FromString parses a GQL SDL string into a set of collection descriptions.
func FromString(ctx context.Context, schemaString string) (
	[]client.CollectionDefinition,
	error,
) {
	source := source.NewSource(&source.Source{
		Body: []byte(schemaString),
	})

	doc, err := gqlp.Parse(
		gqlp.ParseParams{
			Source: source,
		},
	)
	if err != nil {
		return nil, err
	}

	return fromAst(doc)
}

// fromAst parses a GQL AST into a set of collection descriptions.
func fromAst(doc *ast.Document) (
	[]client.CollectionDefinition,
	error,
) {
	definitions := []client.CollectionDefinition{}
	cTypeByFieldNameByObjName := map[string]map[string]client.CType{}

	for _, def := range doc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			description, err := collectionFromAstDefinition(defType, cTypeByFieldNameByObjName)
			if err != nil {
				return nil, err
			}

			definitions = append(definitions, description)

		case *ast.InterfaceDefinition:
			description, err := schemaFromAstDefinition(defType, cTypeByFieldNameByObjName)
			if err != nil {
				return nil, err
			}

			definitions = append(
				definitions,
				client.CollectionDefinition{
					// `Collection` is left as default, as interfaces are schema-only declarations
					Schema: description,
				},
			)

		default:
			// Do nothing, ignore it and continue
			continue
		}
	}

	// The details on the relations between objects depend on both sides
	// of the relationship.  The relation manager handles this, and must be applied
	// after all the collections have been processed.
	err := finalizeRelations(definitions, cTypeByFieldNameByObjName)
	if err != nil {
		return nil, err
	}

	return definitions, nil
}

// collectionFromAstDefinition parses a AST object definition into a set of collection descriptions.
func collectionFromAstDefinition(
	def *ast.ObjectDefinition,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) (client.CollectionDefinition, error) {
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

	indexDescriptions := []client.IndexDescription{}
	for _, field := range def.Fields {
		tmpSchemaFieldDescriptions, tmpCollectionFieldDescriptions, err := fieldsFromAST(
			field,
			def.Name.Value,
			cTypeByFieldNameByObjName,
			false,
		)
		if err != nil {
			return client.CollectionDefinition{}, err
		}

		schemaFieldDescriptions = append(schemaFieldDescriptions, tmpSchemaFieldDescriptions...)
		collectionFieldDescriptions = append(collectionFieldDescriptions, tmpCollectionFieldDescriptions...)

		for _, directive := range field.Directives {
			if directive.Name.Value == types.IndexDirectiveLabel {
				index, err := indexFromAST(directive, field)
				if err != nil {
					return client.CollectionDefinition{}, err
				}
				indexDescriptions = append(indexDescriptions, index)
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

	for _, directive := range def.Directives {
		switch directive.Name.Value {
		case types.IndexDirectiveLabel:
			index, err := indexFromAST(directive, nil)
			if err != nil {
				return client.CollectionDefinition{}, err
			}
			indexDescriptions = append(indexDescriptions, index)

		case types.PolicySchemaDirectiveLabel:
			policy, err := policyFromAST(directive)
			if err != nil {
				return client.CollectionDefinition{}, err
			}
			policyDescription = immutable.Some(policy)
		}
	}

	return client.CollectionDefinition{
		Description: client.CollectionDescription{
			Name:    immutable.Some(def.Name.Value),
			Indexes: indexDescriptions,
			Policy:  policyDescription,
			Fields:  collectionFieldDescriptions,
		},
		Schema: client.SchemaDescription{
			Name:   def.Name.Value,
			Fields: schemaFieldDescriptions,
		},
	}, nil
}

func schemaFromAstDefinition(
	def *ast.InterfaceDefinition,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) (client.SchemaDescription, error) {
	fieldDescriptions := []client.SchemaFieldDescription{}

	for _, field := range def.Fields {
		// schema-only types do not have collection fields, so we can safely discard any returned here
		tmpFieldsDescriptions, _, err := fieldsFromAST(field, def.Name.Value, cTypeByFieldNameByObjName, true)
		if err != nil {
			return client.SchemaDescription{}, err
		}

		fieldDescriptions = append(fieldDescriptions, tmpFieldsDescriptions...)
	}

	// sort the fields lexicographically
	sort.Slice(fieldDescriptions, func(i, j int) bool {
		return fieldDescriptions[i].Name < fieldDescriptions[j].Name
	})

	return client.SchemaDescription{
		Name:   def.Name.Value,
		Fields: fieldDescriptions,
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

func indexFromAST(directive *ast.Directive, fieldDef *ast.FieldDefinition) (client.IndexDescription, error) {
	var name string
	var unique bool

	var direction *ast.EnumValue
	var includes *ast.ListValue

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			nameVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value
			if !IsValidIndexName(name) {
				return client.IndexDescription{}, NewErrIndexWithInvalidName(name)
			}

		case types.IndexDirectivePropIncludes:
			includesVal, ok := arg.Value.(*ast.ListValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			includes = includesVal

		case types.IndexDirectivePropDirection:
			directionVal, ok := arg.Value.(*ast.EnumValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			direction = directionVal

		case types.IndexDirectivePropUnique:
			uniqueVal, ok := arg.Value.(*ast.BooleanValue)
			if !ok {
				return client.IndexDescription{}, ErrIndexWithInvalidArg
			}
			unique = uniqueVal.Value

		default:
			return client.IndexDescription{}, ErrIndexWithUnknownArg
		}
	}

	var containsField bool
	var fields []client.IndexedFieldDescription

	if includes != nil {
		for _, include := range includes.Values {
			field, err := indexFieldFromAST(include, direction)
			if err != nil {
				return client.IndexDescription{}, err
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
		return client.IndexDescription{}, ErrIndexMissingFields
	}

	return client.IndexDescription{
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
		case types.IndexFieldInputName:
			nameVal, ok := field.Value.(*ast.StringValue)
			if !ok {
				return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value

		case types.IndexFieldInputDirection:
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
	var value any
	for _, arg := range directive.Arguments {
		if propName != arg.Name.Value {
			return nil, NewErrDefaultValueInvalid(field.Name.Value, propName, arg.Name.Value)
		}
		switch t := arg.Value.(type) {
		case *ast.IntValue:
			value = gql.Int.ParseLiteral(arg.Value)
		case *ast.FloatValue:
			value = gql.Float.ParseLiteral(arg.Value)
		case *ast.BooleanValue:
			value = t.Value
		case *ast.StringValue:
			value = t.Value
		default:
			value = arg.Value.GetValue()
		}
	}
	return value, nil
}

func fieldsFromAST(
	field *ast.FieldDefinition,
	hostObjectName string,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
	schemaOnly bool,
) ([]client.SchemaFieldDescription, []client.CollectionFieldDescription, error) {
	kind, err := astTypeToKind(field.Type)
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
	for _, directive := range field.Directives {
		if directive.Name.Value == types.DefaultDirectiveLabel {
			defaultValue, err = defaultFromAST(field, directive)
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
			if schemaOnly { // todo - document and/or do better
				schemaFieldDescriptions = append(
					schemaFieldDescriptions,
					client.SchemaFieldDescription{
						Name: field.Name.Value,
						Kind: kind,
					},
				)
			} else {
				collectionFieldDescriptions = append(
					collectionFieldDescriptions,
					client.CollectionFieldDescription{
						Name:         field.Name.Value,
						Kind:         immutable.Some(kind),
						RelationName: immutable.Some(relationName),
					},
				)
			}
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

func astTypeToKind(t ast.Type) (client.FieldKind, error) {
	switch astTypeVal := t.(type) {
	case *ast.List:
		switch innerAstTypeVal := astTypeVal.Type.(type) {
		case *ast.NonNull:
			switch innerAstTypeVal.Type.(*ast.Named).Name.Value {
			case typeBoolean:
				return client.FieldKind_BOOL_ARRAY, nil
			case typeInt:
				return client.FieldKind_INT_ARRAY, nil
			case typeFloat:
				return client.FieldKind_FLOAT_ARRAY, nil
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
			case typeFloat:
				return client.FieldKind_NILLABLE_FLOAT_ARRAY, nil
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
		case typeFloat:
			return client.FieldKind_NILLABLE_FLOAT, nil
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
		return client.FieldKind_None, NewErrTypeNotFound(t.String())
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
	definitions []client.CollectionDefinition,
	cTypeByFieldNameByObjName map[string]map[string]client.CType,
) error {
	embeddedObjNames := map[string]struct{}{}
	for _, def := range definitions {
		if !def.Description.Name.HasValue() {
			embeddedObjNames[def.Schema.Name] = struct{}{}
		}
	}

	for i, definition := range definitions {
		if _, ok := embeddedObjNames[definition.Description.Name.Value()]; ok {
			// Embedded objects are simpler and require no addition work
			continue
		}

		for _, field := range definition.Description.Fields {
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
			for _, otherDef := range definitions {
				// Check the 'other' schema name, there can only be a one-one mapping in an SDL
				// appart from embedded, which will be schema only.
				if otherDef.Schema.Name == namedKind.Name {
					otherColDefinition = immutable.Some(otherDef)
					break
				}
			}

			if !otherColDefinition.HasValue() {
				// If the other collection is not found here we skip this field.  Whilst this almost certainly means the SDL
				// is invalid, validating anything beyond SDL syntax is not the responsibility of this package.
				continue
			}

			otherColFieldDescription, hasOtherColFieldDescription := otherColDefinition.Value().Description.GetFieldByRelation(
				field.RelationName.Value(),
				definition.GetName(),
				field.Name,
			)

			if !hasOtherColFieldDescription || otherColFieldDescription.Kind.Value().IsArray() {
				if _, exists := definition.Schema.GetFieldByName(field.Name); !exists {
					// Relations only defined on one side of the object are possible, and so if this is one of them
					// or if the other side is an array, we need to add the field to the schema (is primary side)
					// if the field has not been explicitly declared by the user.
					definition.Schema.Fields = append(
						definition.Schema.Fields,
						client.SchemaFieldDescription{
							Name: field.Name,
							Kind: field.Kind.Value(),
							Typ:  cTypeByFieldNameByObjName[definition.Schema.Name][field.Name],
						},
					)
				}
			}

			otherIsEmbedded := len(otherColDefinition.Value().Description.Fields) == 0
			if !otherIsEmbedded {
				var schemaFieldIndex int
				var schemaFieldExists bool
				for i, schemaField := range definition.Schema.Fields {
					if schemaField.Name == field.Name {
						schemaFieldIndex = i
						schemaFieldExists = true
						break
					}
				}

				if schemaFieldExists {
					idFieldName := fmt.Sprintf("%s_id", field.Name)

					if _, idFieldExists := definition.Schema.GetFieldByName(idFieldName); !idFieldExists {
						existingFields := definition.Schema.Fields
						definition.Schema.Fields = make([]client.SchemaFieldDescription, len(definition.Schema.Fields)+1)
						copy(definition.Schema.Fields, existingFields[:schemaFieldIndex+1])
						copy(definition.Schema.Fields[schemaFieldIndex+2:], existingFields[schemaFieldIndex+1:])

						// An _id field is added for every 1-1 or 1-N relationship from this object if the relation
						// does not point to an embedded object.
						//
						// It is inserted immediately after the object field to make things nicer for the user.
						definition.Schema.Fields[schemaFieldIndex+1] = client.SchemaFieldDescription{
							Name: idFieldName,
							Kind: client.FieldKind_DocID,
							Typ:  defaultCRDTForFieldKind[client.FieldKind_DocID],
						}
					}
				}
			}

			definitions[i] = definition
		}
	}

	return nil
}
