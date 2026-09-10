// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

import (
	"bytes"
	"embed"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

//go:embed templates/collection.graphql.tmpl
var collectionSchemaTemplates embed.FS

var graphQLNamePattern = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

type collectionTemplateModel struct {
	Collections      []collectionTemplateCollection
	ElementFilters   []collectionTemplateElementFilter
	EncryptedFilters []collectionTemplateElementFilter
}

type collectionTemplateElementFilter struct {
	Name       string
	Scalar     string
	ListType   string
	Comparable bool
	Text       bool
}

type collectionTemplateCollection struct {
	Name             string
	ReadOnly         bool
	Embedded         bool
	HasMutationInput bool
	Fields           []collectionTemplateField
	Numeric          []collectionTemplateField
	InlineArrays     []collectionTemplateField
	NumericArrays    []collectionTemplateField
	RelationMany     []collectionTemplateField
	ForeignIDs       []collectionTemplateField
	Encrypted        []collectionTemplateField
	EmitAVG          bool
	EmitCOUNT        bool
	EmitGROUP        bool
	EmitMAX          bool
	EmitMIN          bool
	EmitSIMILARITY   bool
	EmitSUM          bool
}

type collectionTemplateField struct {
	Name              string
	Type              string
	FilterType        string
	MutationType      string
	OrderType         string
	RelatedName       string
	ElementType       string
	ElementFilterType string
	Relation          bool
	Writable          bool
	Orderable         bool
}

// renderCollectionSchemaSDL renders the collection-dependent portion of the
// public GraphQL schema. It deliberately consumes CollectionVersion instead of
// graphql-go runtime types so it can replace the code-first generator.
func renderCollectionSchemaSDL(collections []client.CollectionVersion) (string, error) {
	return renderCollectionSchemaSDLWithEncryption(collections, false)
}

func renderCollectionSchemaSDLWithEncryption(
	collections []client.CollectionVersion,
	isSearchableEncryptionEnabled bool,
) (string, error) {
	model, err := newCollectionTemplateModel(collections, isSearchableEncryptionEnabled)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New("collection.graphql.tmpl").
		Delims("<%", "%>").
		ParseFS(collectionSchemaTemplates, "templates/collection.graphql.tmpl")
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	if err := tmpl.Execute(&result, model); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.String()) + "\n", nil
}

func renderSchemaSDL(collections []client.CollectionVersion, isSearchableEncryptionEnabled bool) (string, error) {
	dynamic, err := renderCollectionSchemaSDLWithEncryption(collections, isSearchableEncryptionEnabled)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(defaultSchemaSDL) + "\n\n" + dynamic, nil
}

func newCollectionTemplateModel(
	collections []client.CollectionVersion,
	isSearchableEncryptionEnabled bool,
) (collectionTemplateModel, error) {
	byID := make(map[string]string, len(collections))
	bySetRelativeID := make(map[string]string, len(collections))
	for _, collection := range collections {
		byID[collection.CollectionID] = collection.Name
		if collection.CollectionSet.HasValue() {
			set := collection.CollectionSet.Value()
			bySetRelativeID[fmt.Sprintf("%s/%d", set.CollectionSetID, set.RelativeID)] = collection.Name
		}
	}

	result := collectionTemplateModel{Collections: make([]collectionTemplateCollection, 0, len(collections))}
	for _, collection := range collections {
		item := collectionTemplateCollection{
			Name:           collection.Name,
			ReadOnly:       collection.Query.HasValue(),
			Embedded:       collection.IsEmbeddedOnly,
			EmitAVG:        true,
			EmitCOUNT:      true,
			EmitGROUP:      true,
			EmitMAX:        true,
			EmitMIN:        true,
			EmitSIMILARITY: true,
			EmitSUM:        true,
		}
		foreignIDIndexes := map[string]int{}
		for _, field := range collection.Fields {
			if field.Name == request.DocIDFieldName || !strings.HasPrefix(field.Name, "_") {
				continue
			}
			foreignID := collectionTemplateField{
				Name: field.Name, Type: "ID", MutationType: "ID",
				FilterType: "IDOperatorBlock", OrderType: "Ordering", Writable: true,
			}
			item.ForeignIDs = append(item.ForeignIDs, foreignID)
			foreignIDIndexes[field.Name] = len(item.ForeignIDs) - 1
		}
		for _, field := range collection.Fields {
			if field.Name == request.DocIDFieldName || strings.HasPrefix(field.Name, "_") {
				continue
			}
			templateField, foreignID, err := newCollectionTemplateField(
				field, collection, byID, bySetRelativeID,
			)
			if err != nil {
				return collectionTemplateModel{}, fmt.Errorf("collection %s field %s: %w", collection.Name, field.Name, err)
			}
			item.Fields = append(item.Fields, templateField)
			switch field.Name {
			case "AVG":
				item.EmitAVG = false
			case "COUNT":
				item.EmitCOUNT = false
			case "GROUP":
				item.EmitGROUP = false
			case "MAX":
				item.EmitMAX = false
			case "MIN":
				item.EmitMIN = false
			case request.SimilarityFieldName:
				item.EmitSIMILARITY = false
			case "SUM":
				item.EmitSUM = false
			}
			if templateField.Relation {
				if index, exists := foreignIDIndexes["_"+field.Name+"ID"]; exists {
					item.ForeignIDs[index].Writable = field.IsPrimary
				}
			}
			if foreignID != nil {
				if index, exists := foreignIDIndexes[foreignID.Name]; exists {
					item.ForeignIDs[index].Writable = true
				} else {
					item.ForeignIDs = append(item.ForeignIDs, *foreignID)
					foreignIDIndexes[foreignID.Name] = len(item.ForeignIDs) - 1
				}
			}
			if isNumericKind(field.Kind) {
				item.Numeric = append(item.Numeric, templateField)
			}
			if templateField.Relation && field.Kind.IsArray() {
				item.RelationMany = append(item.RelationMany, templateField)
			}
			if _, isArray := field.Kind.(client.ScalarArrayKind); isArray {
				item.InlineArrays = append(item.InlineArrays, templateField)
				if isNumericArrayKind(field.Kind) {
					item.NumericArrays = append(item.NumericArrays, templateField)
				}
			}
		}
		if isSearchableEncryptionEnabled {
			fieldsByName := make(map[string]collectionTemplateField, len(item.Fields))
			for _, field := range item.Fields {
				fieldsByName[field.Name] = field
			}
			for _, index := range collection.EncryptedIndexes {
				field, exists := fieldsByName[index.FieldName]
				if !exists {
					return collectionTemplateModel{}, fmt.Errorf("encrypted index field %s not found", index.FieldName)
				}
				field.FilterType, field.ElementType = encryptedFilterInfo(field.Type)
				item.Encrypted = append(item.Encrypted, field)
			}
		}
		sort.Slice(item.Fields, func(i, j int) bool { return item.Fields[i].Name < item.Fields[j].Name })
		sort.Slice(item.Numeric, func(i, j int) bool { return item.Numeric[i].Name < item.Numeric[j].Name })
		sort.Slice(item.RelationMany, func(i, j int) bool { return item.RelationMany[i].Name < item.RelationMany[j].Name })
		sort.Slice(item.InlineArrays, func(i, j int) bool { return item.InlineArrays[i].Name < item.InlineArrays[j].Name })
		sort.Slice(item.NumericArrays, func(i, j int) bool { return item.NumericArrays[i].Name < item.NumericArrays[j].Name })
		sort.Slice(item.ForeignIDs, func(i, j int) bool { return item.ForeignIDs[i].Name < item.ForeignIDs[j].Name })
		sort.Slice(item.Encrypted, func(i, j int) bool { return item.Encrypted[i].Name < item.Encrypted[j].Name })
		for _, field := range item.ForeignIDs {
			item.HasMutationInput = item.HasMutationInput || field.Writable
		}
		for _, field := range item.Fields {
			item.HasMutationInput = item.HasMutationInput || field.Writable
		}
		result.Collections = append(result.Collections, item)
	}
	sort.Slice(result.Collections, func(i, j int) bool {
		return result.Collections[i].Name < result.Collections[j].Name
	})
	// The code-first generator exposed all scalar element filters for every
	// generated schema, even when the current collections did not contain a
	// matching inline array. Keep that introspection-visible API stable.
	result.ElementFilters = defaultElementFilters()
	filterNames := map[string]struct{}{"IntFilterArg": {}}
	for _, filter := range result.ElementFilters {
		filterNames[filter.Name] = struct{}{}
	}
	for _, collection := range result.Collections {
		for _, field := range collection.InlineArrays {
			if field.ElementFilterType == "JSON" {
				continue
			}
			if _, exists := filterNames[field.ElementFilterType]; exists {
				continue
			}
			filterNames[field.ElementFilterType] = struct{}{}
			scalar := strings.TrimSuffix(field.ElementType, "!")
			result.ElementFilters = append(result.ElementFilters, collectionTemplateElementFilter{
				Name:       field.ElementFilterType,
				Scalar:     scalar,
				ListType:   "[" + field.ElementType + "]",
				Comparable: scalar == "Int" || scalar == "Float32" || scalar == "Float64" || scalar == "DateTime",
				Text:       scalar == "String",
			})
		}
	}
	sort.Slice(result.ElementFilters, func(i, j int) bool {
		return result.ElementFilters[i].Name < result.ElementFilters[j].Name
	})
	encryptedFilterNames := map[string]struct{}{}
	for _, collection := range result.Collections {
		for _, field := range collection.Encrypted {
			if _, exists := encryptedFilterNames[field.FilterType]; exists {
				continue
			}
			encryptedFilterNames[field.FilterType] = struct{}{}
			result.EncryptedFilters = append(result.EncryptedFilters, collectionTemplateElementFilter{
				Name: field.FilterType, Scalar: field.ElementType,
			})
		}
	}
	sort.Slice(result.EncryptedFilters, func(i, j int) bool {
		return result.EncryptedFilters[i].Name < result.EncryptedFilters[j].Name
	})
	return result, nil
}

func defaultElementFilters() []collectionTemplateElementFilter {
	result := make([]collectionTemplateElementFilter, 0, 12)
	for _, scalar := range []struct {
		name       string
		comparable bool
		text       bool
	}{
		{name: "Boolean"},
		{name: "DateTime", comparable: true},
		{name: "Float32", comparable: true},
		{name: "Float64", comparable: true},
		{name: "Int", comparable: true},
		{name: "String", text: true},
	} {
		for _, nonNull := range []bool{false, true} {
			if scalar.name == "Int" && !nonNull {
				// IntFilterArg is part of the static schema because it is used by
				// ScalarAggregateNumericBlock even when no collections exist.
				continue
			}
			prefix := ""
			element := scalar.name
			if nonNull {
				prefix = "NotNull"
				element += "!"
			}
			result = append(result, collectionTemplateElementFilter{
				Name:       prefix + scalar.name + filterInputNameSuffix,
				Scalar:     scalar.name,
				ListType:   "[" + element + "]",
				Comparable: scalar.comparable,
				Text:       scalar.text,
			})
		}
	}
	return result
}

func encryptedFilterInfo(typeName string) (string, string) {
	underlying := strings.Trim(typeName, "[]!")
	prefix := underlying
	if strings.Contains(typeName, "!") {
		prefix = "NotNull" + prefix
	}
	return prefix + encryptedFilterInputNameSuffix, underlying
}

func newCollectionTemplateField(
	field client.CollectionFieldDescription,
	host client.CollectionVersion,
	collectionNamesByID map[string]string,
	collectionNamesBySetRelativeID map[string]string,
) (collectionTemplateField, *collectionTemplateField, error) {
	switch kind := field.Kind.(type) {
	case client.ScalarKind:
		if _, err := strconv.ParseUint(kind.String(), 10, 8); err == nil {
			return collectionTemplateField{}, nil, NewErrTypeNotFound(kind.String())
		}
	case client.ScalarArrayKind:
		if _, err := strconv.ParseUint(kind.String(), 10, 8); err == nil {
			return collectionTemplateField{}, nil, NewErrTypeNotFound(kind.String())
		}
	}
	if !graphQLNamePattern.MatchString(field.Name) {
		return collectionTemplateField{}, nil, fmt.Errorf(
			"Names must match /^[_a-zA-Z][_a-zA-Z0-9]*$/ but %q does not.", field.Name,
		)
	}
	result := collectionTemplateField{Name: field.Name, Writable: true, Orderable: true}
	switch kind := field.Kind.(type) {
	case client.ScalarKind:
		if _, err := strconv.ParseUint(kind.String(), 10, 8); err == nil {
			return collectionTemplateField{}, nil, NewErrTypeNotFound(kind.String())
		}
		result.Type = kind.String()
		result.MutationType = strings.ReplaceAll(kind.String(), "!", "")
		result.FilterType = scalarFilterType(kind.String())
		result.OrderType = "Ordering"
	case client.ScalarArrayKind:
		if _, err := strconv.ParseUint(kind.String(), 10, 8); err == nil {
			return collectionTemplateField{}, nil, NewErrTypeNotFound(kind.String())
		}
		result.Type = kind.String()
		result.MutationType = strings.ReplaceAll(kind.String(), "!", "")
		result.FilterType = scalarListFilterType(kind.String())
		result.OrderType = "Ordering"
		result.ElementType = strings.TrimSuffix(strings.TrimPrefix(kind.String(), "["), "]")
		result.ElementFilterType = scalarArrayElementFilterType(kind.String())
	case *client.NamedKind:
		setRelationTemplateField(&result, kind.Name, kind.Array, field.IsPrimary)
	case *client.CollectionKind:
		name, ok := collectionNamesByID[kind.CollectionID]
		if !ok {
			return collectionTemplateField{}, nil, fmt.Errorf("related collection %q not found", kind.CollectionID)
		}
		setRelationTemplateField(&result, name, kind.Array, field.IsPrimary)
	case *client.SelfKind:
		name := host.Name
		if kind.RelativeID != "" {
			if !host.CollectionSet.HasValue() {
				return collectionTemplateField{}, nil, fmt.Errorf("host has no collection set")
			}
			setID := host.CollectionSet.Value().CollectionSetID
			var ok bool
			name, ok = collectionNamesBySetRelativeID[setID+"/"+kind.RelativeID]
			if !ok {
				return collectionTemplateField{}, nil, fmt.Errorf("relative collection %q not found", kind.RelativeID)
			}
		}
		setRelationTemplateField(&result, name, kind.Array, field.IsPrimary)
	default:
		return collectionTemplateField{}, nil, fmt.Errorf("unsupported field kind %T", field.Kind)
	}

	if !result.Relation || !field.IsPrimary {
		return result, nil, nil
	}
	foreignID := &collectionTemplateField{
		Name:         "_" + field.Name + "ID",
		Type:         "ID",
		MutationType: "ID",
		FilterType:   "IDOperatorBlock",
		Writable:     true,
		OrderType:    "Ordering",
	}
	return result, foreignID, nil
}

func setRelationTemplateField(field *collectionTemplateField, name string, array, primary bool) {
	field.Relation = true
	field.RelatedName = name
	field.Orderable = !array
	field.Type = name
	if array {
		field.Type = "[" + name + "]"
	}
	field.FilterType = name + filterInputNameSuffix
	field.MutationType = "ID"
	field.OrderType = name + "OrderArg"
	field.Writable = primary
}

func scalarFilterType(typeName string) string {
	name := strings.TrimSuffix(typeName, "!")
	if name == "JSON" {
		return name
	}
	return name + "OperatorBlock"
}

func scalarListFilterType(typeName string) string {
	inner := strings.TrimSuffix(strings.TrimPrefix(typeName, "["), "]")
	if inner == "JSON" || inner == "JSON!" {
		return "JSON"
	}
	if strings.HasSuffix(inner, "!") {
		return "NotNull" + strings.TrimSuffix(inner, "!") + "ListOperatorBlock"
	}
	name := inner
	return name + "ListOperatorBlock"
}

func scalarArrayElementFilterType(typeName string) string {
	inner := strings.TrimSuffix(strings.TrimPrefix(typeName, "["), "]")
	if inner == "JSON" || inner == "JSON!" {
		return "JSON"
	}
	if strings.HasSuffix(inner, "!") {
		return "NotNull" + strings.TrimSuffix(inner, "!") + "FilterArg"
	}
	return inner + "FilterArg"
}

func isNumericKind(kind client.FieldKind) bool {
	switch kind {
	case client.FieldKind_INT, client.FieldKind_NILLABLE_INT,
		client.FieldKind_FLOAT32, client.FieldKind_NILLABLE_FLOAT32,
		client.FieldKind_FLOAT64, client.FieldKind_NILLABLE_FLOAT64:
		return true
	default:
		return false
	}
}

func isNumericArrayKind(kind client.FieldKind) bool {
	switch kind {
	case client.FieldKind_INT_ARRAY, client.FieldKind_NILLABLE_INT_ARRAY,
		client.FieldKind_FLOAT32_ARRAY, client.FieldKind_NILLABLE_FLOAT32_ARRAY,
		client.FieldKind_FLOAT64_ARRAY, client.FieldKind_NILLABLE_FLOAT64_ARRAY:
		return true
	default:
		return false
	}
}
