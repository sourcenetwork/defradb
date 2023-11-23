// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import (
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// GeneratePredefinedFromSDL generates documents for GraphQL SDL from a predefined list
// of docs that might include nested docs.
// The SDL is parsed to get the list of fields, and the docs
// are created with the fields parsed from the SDL.
// This allows us to have only one large list of docs with predefined
// fields, and create SDLs with different fields from it.
func GeneratePredefinedFromSDL(gqlSDL string, docsList DocsList) ([]GeneratedDoc, error) {
	resultDocs := make([]GeneratedDoc, 0, len(docsList.Docs))
	typeDefs, err := parseSDL(gqlSDL)
	if err != nil {
		return nil, err
	}
	generator := docGenerator{types: typeDefs}
	for _, doc := range docsList.Docs {
		docs, err := generator.generateRelatedDocs(doc, docsList.ColName)
		if err != nil {
			return nil, err
		}
		resultDocs = append(resultDocs, docs...)
	}
	return resultDocs, nil
}

// GeneratePredefined generates documents from a predefined list
// of docs that might include nested docs.
func GeneratePredefined(defs []client.CollectionDefinition, docsList DocsList) ([]GeneratedDoc, error) {
	resultDocs := make([]GeneratedDoc, 0, len(docsList.Docs))
	typeDefs := make(map[string]client.CollectionDefinition)
	for _, col := range defs {
		typeDefs[col.Description.Name] = col
	}
	generator := docGenerator{types: typeDefs}
	for _, doc := range docsList.Docs {
		docs, err := generator.generateRelatedDocs(doc, docsList.ColName)
		if err != nil {
			return nil, err
		}
		resultDocs = append(resultDocs, docs...)
	}
	return resultDocs, nil
}

type docGenerator struct {
	types map[string]client.CollectionDefinition
}

// toRequestedDoc removes the fields that are not in the SDL of the collection.
func toRequestedDoc(doc map[string]any, typeDef *client.CollectionDefinition) map[string]any {
	result := make(map[string]any)
	for _, field := range typeDef.Schema.Fields {
		if field.IsRelation() || field.Name == request.KeyFieldName {
			continue
		}
		result[field.Name] = doc[field.Name]
	}
	for name, val := range doc {
		if strings.HasSuffix(name, request.RelatedObjectID) {
			result[name] = val
		}
	}
	return result
}

// generatePrimary generates primary docs for the given secondary doc.
func (this *docGenerator) generatePrimary(
	docMap map[string]any,
	typeDef *client.CollectionDefinition,
) (map[string]any, []GeneratedDoc, error) {
	result := []GeneratedDoc{}
	requested := toRequestedDoc(docMap, typeDef)
	for _, field := range typeDef.Schema.Fields {
		if field.IsRelation() {
			if _, hasProp := docMap[field.Name]; hasProp {
				if field.IsPrimaryRelation() {
					subType := this.types[field.Schema]
					subDocMap := toRequestedDoc(docMap[field.Name].(map[string]any), &subType)
					subDoc, err := client.NewDocFromMap(subDocMap)
					if err != nil {
						return nil, nil, NewErrFailedToGenerateDoc(err)
					}
					requested[field.Name+request.RelatedObjectID] = subDoc.Key().String()
					result = append(result, GeneratedDoc{Col: &subType, Doc: subDoc})
				}
			}
		}
	}
	return requested, result, nil
}

// generateRelatedDocs generates related docs (primary and secondary) for the given doc.
func (this *docGenerator) generateRelatedDocs(docMap map[string]any, typeName string) ([]GeneratedDoc, error) {
	typeDef := this.types[typeName]

	requested, result, err := this.generatePrimary(docMap, &typeDef)
	if err != nil {
		return nil, err
	}
	doc, err := client.NewDocFromMap(requested)
	if err != nil {
		return nil, NewErrFailedToGenerateDoc(err)
	}

	result = append(result, GeneratedDoc{Col: &typeDef, Doc: doc})

	var docKey string
	for _, field := range typeDef.Schema.Fields {
		if field.IsRelation() {
			if _, hasProp := docMap[field.Name]; hasProp {
				if !field.IsPrimaryRelation() {
					if docKey == "" {
						docKey = doc.Key().String()
					}
					docs, err := this.generateSecondaryDocs(docMap, typeName, &field, docKey)
					if err != nil {
						return nil, err
					}
					result = append(result, docs...)
				}
			}
		}
	}
	return result, nil
}

// generateSecondaryDocs generates secondary docs for the given primary doc.
func (this *docGenerator) generateSecondaryDocs(
	primaryDoc map[string]any,
	primaryTypeName string,
	relProp *client.FieldDescription,
	primaryDocKey string,
) ([]GeneratedDoc, error) {
	result := []GeneratedDoc{}
	relTypeDef := this.types[relProp.Schema]
	primaryPropName := ""
	for _, relDocProp := range relTypeDef.Schema.Fields {
		if relDocProp.Schema == primaryTypeName && relDocProp.IsPrimaryRelation() {
			primaryPropName = relDocProp.Name + request.RelatedObjectID
			switch relVal := primaryDoc[relProp.Name].(type) {
			case []map[string]any:
				for _, relDoc := range relVal {
					relDoc[primaryPropName] = primaryDocKey
					actions, err := this.generateRelatedDocs(relDoc, relTypeDef.Description.Name)
					if err != nil {
						return nil, err
					}
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocKey
				actions, err := this.generateRelatedDocs(relVal, relTypeDef.Description.Name)
				if err != nil {
					return nil, err
				}
				result = append(result, actions...)
			}
		}
	}
	return result, nil
}
