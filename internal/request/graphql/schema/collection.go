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
	"time"

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
	Name        *collectionName
	Directives  []*collectionDirective
	Fields      []*collectionFieldDefinition
	IsInterface bool
}

// fromAst parses a GQL AST into a set of collection versions.
func fromAst(doc *collectionDocument) (
	[]core.Collection,
	error,
) {
	results := []core.Collection{}

	for _, def := range doc.Definitions {
		result, err := fromAstDefinition(def)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
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
	direction    *collectionEnumValue
	includes     *collectionListValue
	hasUnique    bool
	hasDirection bool
	hasIncludes  bool
}

type indexDirectiveConfig struct {
	name string
	kind string

	// orderedIndexConfig is broken out to support the legacy and newer config
	// options. New and future index types should use the same approach as the
	// `vector` field and just have the single `collectionValue` that is directly
	// parsed.
	ordered orderedIndexConfig
	vector  collectionValue
}

// selectKind records which kind of index the arguments seen so far ask for.
//
// Direction is shared: it is an ordered-index argument, but a vector index has to receive it too so
// that the db layer can reject it with a message about direction rather than about arguments.
func (c *indexDirectiveConfig) selectKind(kind string) error {
	if c.kind != "" && c.kind != kind {
		return ErrIndexWithInvalidArg
	}
	c.kind = kind
	return nil
}

func (c indexDirectiveConfig) newIndex(fieldDef *collectionFieldDefinition) (client.NewIndexRequest, error) {
	switch c.kind {
	case "", types.OrderedIndexKind:
		return orderedIndexFromConfig(c.name, c.ordered, fieldDef)
	case types.VectorIndexKind:
		return vectorIndexFromAST(c.name, c.vector, c.ordered.direction, fieldDef)
	default:
		return client.NewIndexRequest{}, ErrIndexWithInvalidArg
	}
}

func indexFromAST(directive *collectionDirective, fieldDef *collectionFieldDefinition) (client.NewIndexRequest, error) {
	var config indexDirectiveConfig

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.IndexDirectivePropName:
			name, ok := arg.Value.(*collectionStringValue)
			if !ok {
				return client.NewIndexRequest{}, ErrIndexWithInvalidArg
			}
			if !IsValidIndexName(name.Value) {
				return client.NewIndexRequest{}, NewErrIndexWithInvalidName(name.Value)
			}
			config.name = name.Value

		case types.IndexDirectivePropKind:
			kind, ok := arg.Value.(*collectionEnumValue)
			if !ok {
				return client.NewIndexRequest{}, ErrIndexWithInvalidArg
			}
			if err := config.selectKind(kind.Value); err != nil {
				return client.NewIndexRequest{}, err
			}

		case types.OrderedIndexKind:
			if err := config.selectKind(types.OrderedIndexKind); err != nil {
				return client.NewIndexRequest{}, err
			}
			if err := parseOrderedIndexConfig(arg.Value, &config.ordered); err != nil {
				return client.NewIndexRequest{}, err
			}

		case types.VectorIndexKind:
			if err := config.selectKind(types.VectorIndexKind); err != nil {
				return client.NewIndexRequest{}, err
			}
			config.vector = arg.Value

		case types.IndexDirectivePropDirection:
			// Direction does not choose a kind. A vector index cannot honour one, but it has to reach
			// the db layer to be rejected there, where the error can say so.
			if err := parseOrderedIndexProperty(arg.Name.Value, arg.Value, &config.ordered); err != nil {
				return client.NewIndexRequest{}, err
			}

		default:
			if err := config.selectKind(types.OrderedIndexKind); err != nil {
				return client.NewIndexRequest{}, err
			}
			if err := parseOrderedIndexProperty(arg.Name.Value, arg.Value, &config.ordered); err != nil {
				return client.NewIndexRequest{}, err
			}
		}
	}

	return config.newIndex(fieldDef)
}

func parseOrderedIndexConfig(value collectionValue, config *orderedIndexConfig) error {
	obj, ok := value.(*collectionObjectValue)
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

func parseOrderedIndexProperty(name string, value collectionValue, config *orderedIndexConfig) error {
	switch name {
	case types.IndexDirectivePropIncludes:
		if config.hasIncludes {
			return ErrIndexWithInvalidArg
		}
		includes, ok := value.(*collectionListValue)
		if !ok {
			return ErrIndexWithInvalidArg
		}
		config.includes = includes
		config.hasIncludes = true

	case types.IndexDirectivePropDirection:
		if config.hasDirection {
			return ErrIndexWithInvalidArg
		}
		direction, ok := value.(*collectionEnumValue)
		if !ok {
			return ErrIndexWithInvalidArg
		}
		config.direction = direction
		config.hasDirection = true

	case types.IndexDirectivePropUnique:
		if config.hasUnique {
			return ErrIndexWithInvalidArg
		}
		unique, ok := value.(*collectionBooleanValue)
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
	fieldDef *collectionFieldDefinition,
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

func indexFieldFromAST(value collectionValue, defaultDirection *collectionEnumValue) (client.IndexedFieldDescription, error) {
	argTypeObject, ok := value.(*collectionObjectValue)
	if !ok {
		return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
	}

	var name string
	var direction *collectionEnumValue

	for _, field := range argTypeObject.Fields {
		switch field.Name.Value {
		case types.IncludesPropField:
			nameVal, ok := field.Value.(*collectionStringValue)
			if !ok {
				return client.IndexedFieldDescription{}, ErrIndexWithInvalidArg
			}
			name = nameVal.Value

		case types.IncludesPropDirection:
			directionVal, ok := field.Value.(*collectionEnumValue)
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
	field *collectionFieldDefinition,
	directive *collectionDirective,
) (any, error) {
	astNamed, ok := field.Type.(*collectionNamed)
	if !ok {
		// Non-named types (e.g. lists) cannot have a default value.
		return nil, NewErrDefaultValueNotAllowed(field.Name.Value, field.Type.String())
	}
	if len(directive.Arguments) != 1 {
		return nil, NewErrDefaultValueOneArg(field.Name.Value)
	}
	arg := directive.Arguments[0]
	if arg.Name.Value != types.DefaultDirectivePropValue {
		return nil, fmt.Errorf(
			"Unknown argument %q on directive %q", arg.Name.Value, "@default")
	}
	// The value is coerced based on the type of the field the directive is applied to,
	// reusing each scalar's existing ParseLiteral coercion.
	var value any
	switch astNamed.Name.Value {
	case typeInt:
		if literal, ok := arg.Value.(*collectionIntValue); ok {
			parsed, err := strconv.ParseInt(literal.Value, 10, 32)
			if err == nil {
				value = int32(parsed)
			}
		}
	case typeFloat:
		value = parseFloatDefault(arg.Value, 64)
	case typeFloat32:
		value = parseFloatDefault(arg.Value, 32)
	case typeFloat64:
		value = parseFloatDefault(arg.Value, 64)
	case typeBoolean:
		if literal, ok := arg.Value.(*collectionBooleanValue); ok {
			value = literal.Value
		}
	case typeString:
		if literal, ok := arg.Value.(*collectionStringValue); ok {
			value = literal.Value
		}
	case typeDateTime:
		// Handle UTC_NOW as a special case, if that's what the default is
		if enum, ok := arg.Value.(*collectionEnumValue); ok && enum.Value == enum_UTC_NOW {
			value = enum_UTC_NOW
			break
		}
		// Otherwise, parse the value normally as a DateTime
		if literal, ok := arg.Value.(*collectionStringValue); ok {
			parsed, err := time.Parse(time.RFC3339Nano, literal.Value)
			if err == nil {
				value = parsed
			}
		}
	case typeJSON:
		jsonValue := collectionValueToGo(arg.Value)
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
					printLegacyValue(arg.Value),
				)
			}
			value = string(jsonBytes)
		}
	case typeBlob:
		if literal, ok := arg.Value.(*collectionStringValue); ok && types.BlobPattern.MatchString(literal.Value) {
			value = literal.Value
		}
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
			printLegacyValue(arg.Value),
		)
	}
	return value, nil
}

func parseFloatDefault(value collectionValue, bitSize int) any {
	var raw string
	switch value := value.(type) {
	case *collectionIntValue:
		raw = value.Value
	case *collectionFloatValue:
		raw = value.Value
	default:
		return nil
	}
	parsed, err := strconv.ParseFloat(raw, bitSize)
	if err != nil {
		return nil
	}
	if bitSize == 32 {
		return float32(parsed)
	}
	return parsed
}

func collectionValueToGo(value collectionValue) any {
	switch value := value.(type) {
	case *collectionBooleanValue:
		return value.Value
	case *collectionIntValue:
		parsed, err := strconv.ParseInt(value.Value, 10, 32)
		if err == nil {
			return int32(parsed)
		}
	case *collectionFloatValue:
		parsed, err := strconv.ParseFloat(value.Value, 64)
		if err == nil {
			return parsed
		}
	case *collectionStringValue:
		return value.Value
	case *collectionEnumValue:
		return value.Value
	case *collectionNullValue:
		return nil
	case *collectionListValue:
		result := make([]any, len(value.Values))
		for index, item := range value.Values {
			result[index] = collectionValueToGo(item)
		}
		return result
	case *collectionObjectValue:
		result := make(map[string]any, len(value.Fields))
		for _, field := range value.Fields {
			result[field.Name.Value] = collectionValueToGo(field.Value)
		}
		return result
	}
	return nil
}

func defaultValueLiteralType(value collectionValue) string {
	switch value.(type) {
	case *collectionBooleanValue:
		return typeBoolean
	case *collectionIntValue:
		return typeInt
	case *collectionFloatValue:
		return typeFloat
	case *collectionStringValue:
		return typeString
	case *collectionEnumValue:
		return "Enum"
	case *collectionListValue:
		return "List"
	case *collectionObjectValue:
		return "Object"
	case *collectionNullValue:
		return "Null"
	case *collectionVariable:
		return "Variable"
	default:
		return "Unknown"
	}
}

func printLegacyValue(value collectionValue) string {
	switch value := value.(type) {
	case *collectionBooleanValue:
		return strconv.FormatBool(value.Value)
	case *collectionIntValue:
		return value.Value
	case *collectionFloatValue:
		return value.Value
	case *collectionStringValue:
		return strconv.Quote(value.Value)
	case *collectionEnumValue:
		return value.Value
	case *collectionNullValue:
		return "null"
	case *collectionVariable:
		return "$" + value.Name.Value
	case *collectionListValue:
		items := make([]string, len(value.Values))
		for index, item := range value.Values {
			items[index] = printLegacyValue(item)
		}
		return "[" + strings.Join(items, ", ") + "]"
	case *collectionObjectValue:
		fields := make([]string, len(value.Fields))
		for index, field := range value.Fields {
			fields[index] = field.Name.Value + ": " + printLegacyValue(field.Value)
		}
		return "{" + strings.Join(fields, ", ") + "}"
	default:
		return ""
	}
}

func encryptedIndexFromAST(
	directive *collectionDirective,
	fieldDef *collectionFieldDefinition,
) (client.EncryptedIndexDescription, error) {
	encryptedIndex := client.EncryptedIndexDescription{
		FieldName: fieldDef.Name.Value,
		Type:      client.EncryptedIndexTypeEquality,
	}

	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.EncryptedIndexDirectivePropType:
			typeVal, ok := arg.Value.(*collectionStringValue)
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
	field *collectionFieldDefinition,
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
func policyFromAST(directive *collectionDirective) (client.PolicyDescription, error) {
	policyDesc := client.PolicyDescription{}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.PolicySchemaDirectivePropID:
			policyIDProp, ok := arg.Value.(*collectionStringValue)
			if !ok {
				return client.PolicyDescription{}, fmt.Errorf(
					"Argument %q has invalid value %v", arg.Name.Value, arg.Value.GetValue())
			}
			policyDesc.ID = policyIDProp.Value
		case types.PolicySchemaDirectivePropResource:
			policyResourceProp, ok := arg.Value.(*collectionStringValue)
			if !ok {
				return client.PolicyDescription{}, fmt.Errorf(
					"Argument %q has invalid value %v", arg.Name.Value, arg.Value.GetValue())
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
	config collectionValue,
	direction *collectionEnumValue,
	fieldDef *collectionFieldDefinition,
) (client.NewIndexRequest, error) {
	if fieldDef == nil {
		return client.NewIndexRequest{}, ErrIndexWithInvalidArg
	}

	obj, ok := config.(*collectionObjectValue)
	if !ok {
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
	for _, field := range obj.Fields {
		switch field.Name.Value {
		case types.VectorIndexPropDimensions:
			parsed, err := parseUint32ASTValue(field.Value)
			if err != nil {
				return client.NewIndexRequest{}, err
			}
			dimensions = parsed

		case types.VectorIndexPropAlgorithm:
			algorithmVal, ok := field.Value.(*collectionEnumValue)
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
			return client.NewIndexRequest{}, fmt.Errorf(
				"In field %q: Unknown field.: %w", field.Name.Value, ErrIndexWithUnknownArg)
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

	field := client.IndexedFieldDescription{Name: fieldDef.Name.Value}
	if direction != nil {
		field.Descending = direction.Value == types.FieldOrderDESC
	}

	return client.NewIndexRequest{
		Name:   name,
		Fields: []client.IndexedFieldDescription{field},
		Vector: &vectorDesc,
	}, nil
}

// parseHNSWConfig reads the @index vector HNSW config, overwriting only explicitly set defaults.
func parseHNSWConfig(value collectionValue, metric *client.DistanceMetric, params *client.HNSWParams) error {
	obj, ok := value.(*collectionObjectValue)
	if !ok {
		return ErrIndexWithInvalidArg
	}

	for _, field := range obj.Fields {
		switch field.Name.Value {
		case types.VectorIndexConfigPropMetric:
			metricVal, ok := field.Value.(*collectionEnumValue)
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
				return fmt.Errorf(
					"Expected type %q, found %s: %w",
					"VectorDistanceMetric", metricVal.Value, NewErrVectorIndexUnknownMetric(metricVal.Value))
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
func parseUint32ASTValue(value collectionValue) (uint32, error) {
	intVal, ok := value.(*collectionIntValue)
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
	directive *collectionDirective,
	fieldDef *collectionFieldDefinition,
) (client.VectorEmbeddingDescription, error) {
	embedding := client.VectorEmbeddingDescription{
		FieldName: fieldDef.Name.Value,
	}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.VectorEmbeddingDirectivePropFields:
			val := arg.Value.(*collectionListValue)
			fields := make([]string, len(val.Values))
			for i, untypedField := range val.Values {
				fields[i] = untypedField.(*collectionStringValue).Value
			}
			embedding.Fields = fields
		case types.VectorEmbeddingDirectivePropModel:
			embedding.Model = arg.Value.(*collectionStringValue).Value
		case types.VectorEmbeddingDirectivePropProvider:
			embedding.Provider = arg.Value.(*collectionStringValue).Value
		case types.VectorEmbeddingDirectivePropTemplate:
			embedding.Template = arg.Value.(*collectionStringValue).Value
		case types.VectorEmbeddingDirectivePropURL:
			embedding.URL = arg.Value.(*collectionStringValue).Value
		}
	}
	return embedding, nil
}

type constraintDescription struct {
	Size int
}

func constraintsFromAST(kind client.FieldKind, directive *collectionDirective) (constraintDescription, error) {
	constraints := constraintDescription{}
	for _, arg := range directive.Arguments {
		switch arg.Name.Value {
		case types.ConstraintsDirectivePropSize:
			if !kind.IsArray() {
				return constraintDescription{}, NewErrInvalidTypeForContraint(kind)
			}
			size, err := strconv.Atoi(arg.Value.(*collectionIntValue).Value)
			if err != nil {
				return constraintDescription{}, err
			}
			constraints.Size = size
		}
	}
	return constraints, nil
}

func setCRDTType(field *collectionFieldDefinition, kind client.FieldKind) (client.CType, error) {
	if directive, exists := findDirective(field, "crdt"); exists {
		for _, arg := range directive.Arguments {
			switch arg.Name.Value {
			case "type":
				if stringValue, ok := arg.Value.(*collectionStringValue); ok {
					return 0, fmt.Errorf(
						"Argument %q has invalid value %q", arg.Name.Value, stringValue.Value)
				}
				cTypeString := arg.Value.GetValue().(string)
				cType, validCRDTEnum := types.ParseCRDTType(cTypeString)
				if !validCRDTEnum {
					return 0, client.NewErrInvalidCRDTType(field.Name.Value, cTypeString)
				}
				return cType, nil
			}
		}
	}

	return defaultCRDTForFieldKind[kind], nil
}

func astTypeToKind(
	hostObjectName string,
	field *collectionFieldDefinition,
) (client.FieldKind, error) {
	switch astTypeVal := field.Type.(type) {
	case *collectionList:
		if isNestedListType(astTypeVal.Type) {
			return client.FieldKind_None, NewErrNestedListTypeNotSupported(hostObjectName, field.Name.Value)
		}

		switch innerAstTypeVal := astTypeVal.Type.(type) {
		case *collectionNonNull:
			switch innerAstTypeVal.Type.(*collectionNamed).Name.Value {
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
				return client.FieldKind_None, NewErrNonNullForTypeNotSupported(innerAstTypeVal.Type.(*collectionNamed).Name.Value)
			}

		default:
			switch astTypeVal.Type.(*collectionNamed).Name.Value {
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
				return client.NewNamedKind(astTypeVal.Type.(*collectionNamed).Name.Value, true), nil
			}
		}

	case *collectionNamed:
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

	case *collectionNonNull:
		namedType, ok := astTypeVal.Type.(*collectionNamed)
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

func isNestedListType(fieldType collectionType) bool {
	switch typeVal := fieldType.(type) {
	case *collectionList:
		return true
	case *collectionNonNull:
		_, isList := typeVal.Type.(*collectionList)
		return isList
	default:
		return false
	}
}

func findDirective(field *collectionFieldDefinition, directiveName string) (*collectionDirective, bool) {
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
	field *collectionFieldDefinition,
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
