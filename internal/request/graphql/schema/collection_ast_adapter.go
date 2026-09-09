// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

import (
	"strconv"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
)

func adaptCollectionDocument(document *wgast.Document) *collectionDocument {
	result := &collectionDocument{}
	for _, root := range document.RootNodes {
		var item *typeDefinition
		switch root.Kind {
		case wgast.NodeKindObjectTypeDefinition:
			definition := document.ObjectTypeDefinitions[root.Ref]
			item = &typeDefinition{
				Name:       collectionNameOf(document.ObjectTypeDefinitionNameString(root.Ref)),
				Directives: adaptDirectives(document, definition.Directives.Refs),
				Fields:     adaptFields(document, definition.FieldsDefinition.Refs),
			}
		case wgast.NodeKindInterfaceTypeDefinition:
			definition := document.InterfaceTypeDefinitions[root.Ref]
			item = &typeDefinition{
				Name:        collectionNameOf(document.InterfaceTypeDefinitionNameString(root.Ref)),
				Directives:  adaptDirectives(document, definition.Directives.Refs),
				Fields:      adaptFields(document, definition.FieldsDefinition.Refs),
				IsInterface: true,
			}
		}
		if item != nil {
			result.Definitions = append(result.Definitions, item)
		}
	}
	return result
}

func adaptFields(document *wgast.Document, refs []int) []*collectionFieldDefinition {
	result := make([]*collectionFieldDefinition, 0, len(refs))
	for _, ref := range refs {
		result = append(result, &collectionFieldDefinition{
			Name:       collectionNameOf(document.FieldDefinitionNameString(ref)),
			Type:       adaptType(document, document.FieldDefinitionType(ref)),
			Directives: adaptDirectives(document, document.FieldDefinitionDirectives(ref)),
		})
	}
	return result
}

func adaptDirectives(document *wgast.Document, refs []int) []*collectionDirective {
	result := make([]*collectionDirective, 0, len(refs))
	for _, ref := range refs {
		arguments := make([]*collectionArgument, 0, len(document.DirectiveArgumentSet(ref)))
		for _, argumentRef := range document.DirectiveArgumentSet(ref) {
			arguments = append(arguments, &collectionArgument{
				Name:  collectionNameOf(document.ArgumentNameString(argumentRef)),
				Value: adaptValue(document, document.ArgumentValue(argumentRef)),
			})
		}
		result = append(result, &collectionDirective{
			Name:      collectionNameOf(document.DirectiveNameString(ref)),
			Arguments: arguments,
		})
	}
	return result
}

func adaptType(document *wgast.Document, ref int) collectionType {
	typeRef := document.Types[ref]
	switch typeRef.TypeKind {
	case wgast.TypeKindNamed:
		return &collectionNamed{Name: collectionNameOf(document.TypeNameString(ref))}
	case wgast.TypeKindList:
		return &collectionList{Type: adaptType(document, typeRef.OfType)}
	case wgast.TypeKindNonNull:
		return &collectionNonNull{Type: adaptType(document, typeRef.OfType)}
	}
	return nil
}

func adaptValue(document *wgast.Document, value wgast.Value) collectionValue {
	switch value.Kind {
	case wgast.ValueKindString:
		content := document.ValueContentString(value)
		if decoded, err := strconv.Unquote(`"` + content + `"`); err == nil {
			content = decoded
		}
		return &collectionStringValue{Value: content}
	case wgast.ValueKindBoolean:
		return &collectionBooleanValue{Value: bool(document.BooleanValues[value.Ref])}
	case wgast.ValueKindInteger:
		return &collectionIntValue{Value: signedValue(
			document.IntValueIsNegative(value.Ref), string(document.IntValueRaw(value.Ref)))}
	case wgast.ValueKindFloat:
		return &collectionFloatValue{Value: signedValue(
			document.FloatValueIsNegative(value.Ref), string(document.FloatValueRaw(value.Ref)))}
	case wgast.ValueKindEnum:
		return &collectionEnumValue{Value: document.ValueContentString(value)}
	case wgast.ValueKindNull:
		return &collectionNullValue{}
	case wgast.ValueKindVariable:
		return &collectionVariable{Name: collectionNameOf(document.VariableValueNameString(value.Ref))}
	case wgast.ValueKindList:
		values := make([]collectionValue, 0, len(document.ListValues[value.Ref].Refs))
		for _, ref := range document.ListValues[value.Ref].Refs {
			values = append(values, adaptValue(document, document.Value(ref)))
		}
		return &collectionListValue{Values: values}
	case wgast.ValueKindObject:
		fields := make([]*collectionObjectField, 0, len(document.ObjectValues[value.Ref].Refs))
		for _, ref := range document.ObjectValues[value.Ref].Refs {
			fields = append(fields, &collectionObjectField{
				Name:  collectionNameOf(document.ObjectFieldNameString(ref)),
				Value: adaptValue(document, document.ObjectFieldValue(ref)),
			})
		}
		return &collectionObjectValue{Fields: fields}
	}
	return nil
}

func collectionNameOf(value string) *collectionName { return &collectionName{Value: value} }

func signedValue(negative bool, value string) string {
	if negative {
		return "-" + value
	}
	return value
}
