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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/graphql-go/language/printer"
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

	// Special case enums
	enum_UTC_NOW string = "UTC_NOW"
)

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

	for _, def := range doc.Definitions {
		switch defType := def.(type) {
		case *ast.ObjectDefinition:
			td := newObjectDefinition(defType)
			result, err := fromAstDefinition(td)
			if err != nil {
				return nil, err
			}

			results = append(results, result)

		case *ast.InterfaceDefinition:
			td := newInterfaceDefinition(defType)
			result, err := fromAstDefinition(td)
			if err != nil {
				return nil, err
			}

			results = append(results, result)

		default:
			// Do nothing, ignore it and continue
			continue
		}
	}

	return results, nil
}

// fromAstDefinition parses a AST object definition into a set of collection versions.
func fromAstDefinition(
	def *typeDefinition,
) (core.Collection, error) {
	collectionFieldDescriptions := []client.CollectionFieldDescription{
		{
			Name: request.DocIDFieldName,
			Kind: client.FieldKind_DocID,
			Typ:  client.NONE_CRDT,
		},
	}

	policyDescription := immutable.None[client.PolicyDescription]()

	indexes := []client.NewIndexRequest{}
	vectorEmbeddings := []client.VectorEmbeddingDescription{}
	encryptedIndexes := []client.EncryptedIndexDescription{}
	for _, field := range def.Fields {
		tmpCollectionFieldDescriptions, err := fieldsFromAST(
			field,
			def.Name.Value,
		)
		if err != nil {
			return core.Collection{}, err
		}

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
		Definition: client.CollectionVersion{
			Name:             def.Name.Value,
			Policy:           policyDescription,
			Fields:           collectionFieldDescriptions,
			IsMaterialized:   !isMaterialized.HasValue() || isMaterialized.Value(),
			IsBranchable:     isBranchable,
			IsEmbeddedOnly:   def.IsInterface,
			IsActive:         true,
			VectorEmbeddings: vectorEmbeddings,
			EncryptedIndexes: encryptedIndexes,
		},
		NewIndexes: indexes,
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

type orderedIndexConfig struct {
	unique       bool
	direction    *ast.EnumValue
	includes     *ast.ListValue
	hasUnique    bool
	hasDirection bool
	hasIncludes  bool
}

func indexFromAST(directive *ast.Directive, fieldDef *ast.FieldDefinition) (client.NewIndexRequest, error) {
	var name string
	var kind string
	var orderedConfig orderedIndexConfig
	var vectorConfig ast.Value
	var hasLegacyOrderedConfig bool
	var hasOrderedConfig bool
	var hasVectorConfig bool

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			nameVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.NewIndexRequest{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value
			if !IsValidIndexName(name) {
				return client.NewIndexRequest{}, NewErrIndexWithInvalidName(name)
			}

		case types.IndexDirectivePropKind:
			kindVal, ok := arg.Value.(*ast.EnumValue)
			if !ok {
				return client.NewIndexRequest{}, ErrIndexWithInvalidArg
			}
			kind = kindVal.Value

		case types.OrderedIndexKind:
			hasOrderedConfig = true
			if err := parseOrderedIndexConfig(arg.Value, &orderedConfig); err != nil {
				return client.NewIndexRequest{}, err
			}

		case types.VectorIndexKind:
			hasVectorConfig = true
			vectorConfig = arg.Value

		default:
			hasLegacyOrderedConfig = true
			if err := parseOrderedIndexProperty(arg.Name.Value, arg.Value, &orderedConfig); err != nil {
				return client.NewIndexRequest{}, err
			}
		}
	}

	if hasOrderedConfig && (hasVectorConfig || hasLegacyOrderedConfig) ||
		hasVectorConfig && hasLegacyOrderedConfig ||
		kind == types.OrderedIndexKind && hasVectorConfig ||
		kind == types.VectorIndexKind && (hasOrderedConfig || hasLegacyOrderedConfig) {
		return client.NewIndexRequest{}, ErrIndexWithInvalidArg
	}

	selectedKind := kind
	if selectedKind == "" {
		if hasVectorConfig {
			selectedKind = types.VectorIndexKind
		} else {
			selectedKind = types.OrderedIndexKind
		}
	}

	switch selectedKind {
	case types.OrderedIndexKind:
		return orderedIndexFromConfig(name, orderedConfig, fieldDef)
	case types.VectorIndexKind:
		return vectorIndexFromAST(name, vectorConfig, fieldDef)
	default:
		return client.NewIndexRequest{}, ErrIndexWithInvalidArg
	}
}

func parseOrderedIndexConfig(value ast.Value, config *orderedIndexConfig) error {
	obj, ok := value.(*ast.ObjectValue)
	if !ok {
		return ErrIndexWithInvalidArg
	}
	for _, field := range obj.Fields {
		if err := parseOrderedIndexProperty(field.Name.Value, field.Value, config); err != nil {
			return err
		}
	}
	return nil
}

func parseOrderedIndexProperty(name string, value ast.Value, config *orderedIndexConfig) error {
	switch name {
	case types.IndexDirectivePropIncludes:
		if config.hasIncludes {
			return ErrIndexWithInvalidArg
		}
		includes, ok := value.(*ast.ListValue)
		if !ok {
			return ErrIndexWithInvalidArg
		}
		config.includes = includes
		config.hasIncludes = true

	case types.IndexDirectivePropDirection:
		if config.hasDirection {
			return ErrIndexWithInvalidArg
		}
		direction, ok := value.(*ast.EnumValue)
		if !ok {
			return ErrIndexWithInvalidArg
		}
		config.direction = direction
		config.hasDirection = true

	case types.IndexDirectivePropUnique:
		if config.hasUnique {
			return ErrIndexWithInvalidArg
		}
		unique, ok := value.(*ast.BooleanValue)
		if !ok {
			return ErrIndexWithInvalidArg
		}
		config.unique = unique.Value
		config.hasUnique = true

	default:
		return ErrIndexWithUnknownArg
	}
	return nil
}

func orderedIndexFromConfig(
	name string,
	config orderedIndexConfig,
	fieldDef *ast.FieldDefinition,
) (client.NewIndexRequest, error) {
	var containsField bool
	var fields []client.IndexedFieldDescription

	if config.includes != nil {
		for _, include := range config.includes.Values {
			field, err := indexFieldFromAST(include, config.direction)
			if err != nil {
				return client.NewIndexRequest{}, err
			}
			if fieldDef != nil && fieldDef.Name.Value == field.Name {
				containsField = true
			}
			fields = append(fields, field)
		}
	}

	// If the directive is applied to a field that is not in the includes list, add it first.
	if !containsField && fieldDef != nil {
		field := client.IndexedFieldDescription{Name: fieldDef.Name.Value}
		if config.direction != nil {
			field.Descending = config.direction.Value == types.FieldOrderDESC
		}
		fields = append([]client.IndexedFieldDescription{field}, fields...)
	}

	if len(fields) == 0 {
		return client.NewIndexRequest{}, ErrIndexMissingFields
	}

	return client.NewIndexRequest{
		Name:   name,
		Fields: fields,
		Unique: config.unique,
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
		// Non-named types (e.g. lists) cannot have a default value.
		return nil, NewErrDefaultValueNotAllowed(field.Name.Value, field.Type.String())
	}
	if len(directive.Arguments) != 1 {
		return nil, NewErrDefaultValueOneArg(field.Name.Value)
	}
	arg := directive.Arguments[0]
	if arg.Name.Value != types.DefaultDirectivePropValue {
		// Defensive: GraphQL validation (KnownArgumentNamesRule) already rejects any
		// argument other than `value`, but guard anyway.
		return nil, NewErrDefaultValueOneArg(field.Name.Value)
	}
	// The value is coerced based on the type of the field the directive is applied to,
	// reusing each scalar's existing ParseLiteral coercion.
	var value any
	switch astNamed.Name.Value {
	case typeInt:
		value = gql.Int.ParseLiteral(arg.Value, nil)
	case typeFloat:
		value = gql.Float.ParseLiteral(arg.Value, nil)
	case typeFloat32:
		value = types.Float32.ParseLiteral(arg.Value, nil)
	case typeFloat64:
		value = types.Float64.ParseLiteral(arg.Value, nil)
	case typeBoolean:
		value = gql.Boolean.ParseLiteral(arg.Value, nil)
	case typeString:
		value = gql.String.ParseLiteral(arg.Value, nil)
	case typeDateTime:
		// Handle UTC_NOW as a special case, if that's what the default is
		if enum, ok := arg.Value.(*ast.EnumValue); ok && enum.Value == enum_UTC_NOW {
			value = enum_UTC_NOW
			break
		}
		// Otherwise, parse the value normally as a DateTime
		value = gql.DateTime.ParseLiteral(arg.Value, nil)
	case typeJSON:
		jsonValue := types.JSON.ParseLiteral(arg.Value, nil)
		switch v := jsonValue.(type) {
		case nil:
			value = nil
		case string, int32, float64, bool:
			value = v
		default:
			// If the value is not a primitive type, marshal it to a JSON string for storage
			jsonBytes, err := json.Marshal(jsonValue)
			if err != nil {
				return nil, NewErrDefaultValueInvalid(
					field.Name.Value,
					astNamed.Name.Value,
					defaultValueLiteralType(arg.Value),
					printer.Print(arg.Value),
				)
			}
			value = string(jsonBytes)
		}
	case typeBlob:
		value = types.Blob.ParseLiteral(arg.Value, nil)
	default:
		// Field types not present above (e.g. ID, relations) cannot have a default value.
		return nil, NewErrDefaultValueNotAllowed(field.Name.Value, astNamed.Name.Value)
	}
	// If the value is nil, then parsing has failed, or a nil value was provided.
	// Since setting a default value to nil is the same as not providing one,
	// it is safer to return an error to let the user know something is wrong.
	if value == nil {
		return nil, NewErrDefaultValueInvalid(
			field.Name.Value,
			astNamed.Name.Value,
			defaultValueLiteralType(arg.Value),
			printer.Print(arg.Value),
		)
	}
	return value, nil
}

func defaultValueLiteralType(value ast.Value) string {
	switch value.(type) {
	case *ast.BooleanValue:
		return typeBoolean
	case *ast.IntValue:
		return typeInt
	case *ast.FloatValue:
		return typeFloat
	case *ast.StringValue:
		return typeString
	case *ast.EnumValue:
		return "Enum"
	case *ast.ListValue:
		return "List"
	case *ast.ObjectValue:
		return "Object"
	case *ast.NullValue:
		return "Null"
	case *ast.Variable:
		return "Variable"
	default:
		return "Unknown"
	}
}

func encryptedIndexFromAST(
	directive *ast.Directive,
	fieldDef *ast.FieldDefinition,
) (client.EncryptedIndexDescription, error) {
	encryptedIndex := client.EncryptedIndexDescription{
		FieldName: fieldDef.Name.Value,
		Type:      client.EncryptedIndexTypeEquality,
	}

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.EncryptedIndexDirectivePropType:
			typeVal, ok := arg.Value.(*ast.StringValue)
			if !ok {
				return client.EncryptedIndexDescription{}, NewErrEncryptedIndexWithInvalidArg(fieldDef.Name.Value)
			}

			// Currently only equality is supported
			if typeVal.Value != string(client.EncryptedIndexTypeEquality) {
				return client.EncryptedIndexDescription{}, NewErrEncryptedIndexTypeNotSupported(typeVal.Value)
			}
			encryptedIndex.Type = client.EncryptedIndexType(typeVal.Value)

		default:
			return client.EncryptedIndexDescription{}, NewErrEncryptedIndexWithUnknownArg(arg.Name.Value)
		}
	}

	return encryptedIndex, nil
}

func fieldsFromAST(
	field *ast.FieldDefinition,
	hostObjectName string,
) ([]client.CollectionFieldDescription, error) {
	kind, err := astTypeToKind(hostObjectName, field)
	if err != nil {
		return nil, err
	}

	cType, err := setCRDTType(field, kind)
	if err != nil {
		return nil, err
	}

	var defaultValue any
	var constraints constraintDescription
	for _, directive := range field.Directives {
		switch directive.Name.Value {
		case types.DefaultDirectiveLabel:
			defaultValue, err = defaultFromAST(field, directive)
			if err != nil {
				return nil, err
			}
		case types.ConstraintsDirectiveLabel:
			constraints, err = constraintsFromAST(kind, directive)
			if err != nil {
				return nil, err
			}
		}
	}

	collectionFieldDescriptions := []client.CollectionFieldDescription{}

	if namedKind, ok := kind.(*client.NamedKind); ok {
		relationName, err := getRelationshipName(field, hostObjectName, namedKind.Name)
		if err != nil {
			return nil, err
		}

		if kind.IsArray() {
			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         field.Name.Value,
					Kind:         kind,
					RelationName: immutable.Some(relationName),
				},
			)
		} else {
			idFieldName := request.ToFieldID(field.Name.Value)
			_, isPrimary := findDirective(field, "primary")

			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         idFieldName,
					Kind:         client.FieldKind_DocID,
					Typ:          client.LWW_REGISTER,
					IsPrimary:    isPrimary,
					RelationName: immutable.Some(relationName),
				},
			)

			collectionFieldDescriptions = append(
				collectionFieldDescriptions,
				client.CollectionFieldDescription{
					Name:         field.Name.Value,
					Kind:         kind,
					IsPrimary:    isPrimary,
					RelationName: immutable.Some(relationName),
				},
			)
		}
	} else {
		collectionFieldDescriptions = append(
			collectionFieldDescriptions,
			client.CollectionFieldDescription{
				Name:         field.Name.Value,
				Kind:         kind,
				Typ:          cType,
				DefaultValue: defaultValue,
				Size:         constraints.Size,
			},
		)
	}

	return collectionFieldDescriptions, nil
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

func vectorIndexFromAST(
	name string,
	config ast.Value,
	fieldDef *ast.FieldDefinition,
) (client.NewIndexRequest, error) {
	if fieldDef == nil {
		return client.NewIndexRequest{}, ErrIndexWithInvalidArg
	}

	var dimensions uint32
	algorithm := client.VectorAlgorithmHNSW
	metric := client.DistanceMetricCosine
	hnswParams := client.HNSWParams{
		M:              client.DefaultHNSWM,
		EfConstruction: client.DefaultHNSWEfConstruction,
		EfSearch:       client.DefaultHNSWEfSearch,
	}

	if config != nil {
		obj, ok := config.(*ast.ObjectValue)
		if !ok {
			return client.NewIndexRequest{}, ErrIndexWithInvalidArg
		}
		for _, field := range obj.Fields {
			switch field.Name.Value {
			case types.VectorIndexPropDimensions:
				parsed, err := parseUint32ASTValue(field.Value)
				if err != nil {
					return client.NewIndexRequest{}, err
				}
				dimensions = parsed

			case types.VectorIndexPropAlgorithm:
				algorithmVal, ok := field.Value.(*ast.EnumValue)
				if !ok || algorithmVal.Value != types.VectorIndexAlgorithmHNSW {
					return client.NewIndexRequest{}, ErrIndexWithInvalidArg
				}
				algorithm = client.VectorAlgorithmHNSW

			case types.VectorIndexPropHNSW:
				algorithm = client.VectorAlgorithmHNSW
				if err := parseHNSWConfig(field.Value, &metric, &hnswParams); err != nil {
					return client.NewIndexRequest{}, err
				}

			default:
				return client.NewIndexRequest{}, ErrIndexWithUnknownArg
			}
		}
	}

	vectorDesc := client.VectorIndexDescription{
		Algorithm:  algorithm,
		Metric:     metric,
		Dimensions: dimensions,
	}
	if algorithm == client.VectorAlgorithmHNSW {
		vectorDesc.HNSW = &hnswParams
	}

	return client.NewIndexRequest{
		Name:   name,
		Fields: []client.IndexedFieldDescription{{Name: fieldDef.Name.Value}},
		Vector: &vectorDesc,
	}, nil
}

// parseHNSWConfig reads the @index vector HNSW config, overwriting only explicitly set defaults.
func parseHNSWConfig(value ast.Value, metric *client.DistanceMetric, params *client.HNSWParams) error {
	obj, ok := value.(*ast.ObjectValue)
	if !ok {
		return ErrIndexWithInvalidArg
	}

	for _, field := range obj.Fields {
		switch field.Name.Value {
		case types.VectorIndexConfigPropMetric:
			metricVal, ok := field.Value.(*ast.EnumValue)
			if !ok {
				return ErrIndexWithInvalidArg
			}
			switch metricVal.Value {
			case types.VectorDistanceMetricCosine:
				*metric = client.DistanceMetricCosine
			case types.VectorDistanceMetricEuclidean:
				*metric = client.DistanceMetricEuclidean
			case types.VectorDistanceMetricDot:
				*metric = client.DistanceMetricDotProduct
			default:
				return NewErrVectorIndexUnknownMetric(metricVal.Value)
			}

		case types.VectorIndexHNSWConfigPropM:
			parsed, err := parseUint32ASTValue(field.Value)
			if err != nil {
				return err
			}
			params.M = parsed

		case types.VectorIndexHNSWConfigPropEfConstruction:
			parsed, err := parseUint32ASTValue(field.Value)
			if err != nil {
				return err
			}
			params.EfConstruction = parsed

		case types.VectorIndexHNSWConfigPropEfSearch:
			parsed, err := parseUint32ASTValue(field.Value)
			if err != nil {
				return err
			}
			params.EfSearch = parsed

		default:
			return ErrIndexWithUnknownArg
		}
	}
	return nil
}

// parseUint32ASTValue reads an AST int literal into a uint32, rejecting non-ints and out-of-range
// values.
func parseUint32ASTValue(value ast.Value) (uint32, error) {
	intVal, ok := value.(*ast.IntValue)
	if !ok {
		return 0, ErrIndexWithInvalidArg
	}
	parsed, err := strconv.ParseUint(intVal.Value, 10, 32)
	if err != nil {
		return 0, ErrIndexWithInvalidArg
	}
	return uint32(parsed), nil
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

func constraintsFromAST(kind client.FieldKind, directive *ast.Directive) (constraintDescription, error) {
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
		if isNestedListType(astTypeVal.Type) {
			return client.FieldKind_None, NewErrNestedListTypeNotSupported(hostObjectName, field.Name.Value)
		}

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
			case typeDateTime:
				return client.FieldKind_DATETIME_ARRAY, nil
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
			case typeDateTime:
				return client.FieldKind_NILLABLE_DATETIME_ARRAY, nil
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
		namedType, ok := astTypeVal.Type.(*ast.Named)
		if !ok {
			return client.FieldKind_None, ErrNonNullNotSupported
		}
		switch namedType.Name.Value {
		case typeBoolean:
			return client.FieldKind_BOOL, nil
		case typeInt:
			return client.FieldKind_INT, nil
		case typeFloat, typeFloat64:
			return client.FieldKind_FLOAT64, nil
		case typeFloat32:
			return client.FieldKind_FLOAT32, nil
		case typeDateTime:
			return client.FieldKind_DATETIME, nil
		case typeString:
			return client.FieldKind_STRING, nil
		case typeBlob:
			return client.FieldKind_BLOB, nil
		case typeJSON:
			return client.FieldKind_JSON, nil
		default:
			return client.FieldKind_None, ErrNonNullNotSupported
		}

	default:
		if field.Type == nil {
			return client.FieldKind_None, NewErrFieldTypeNotSpecified(hostObjectName, field.Name.Value)
		}
		return client.FieldKind_None, NewErrTypeNotFound(field.Type.String())
	}
}

func isNestedListType(fieldType ast.Type) bool {
	switch typeVal := fieldType.(type) {
	case *ast.List:
		return true
	case *ast.NonNull:
		_, isList := typeVal.Type.(*ast.List)
		return isList
	default:
		return false
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
