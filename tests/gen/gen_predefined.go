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
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// GeneratePredefinedFromSDL generates documents for a schema from a predefined list
// of docs that might include nested docs.
// The schema is parsed to get the list of fields, and the docs
// are created with the fields parsed from the schema.
// This allows us to have only one large list of docs with predefined
// fields, and create schemas with different fields from it.
func GeneratePredefinedFromSDL(gqlSDL string, docsList DocsList) ([]GeneratedDoc, error) {
	resultDocs := make([]GeneratedDoc, 0, len(docsList.Docs))
	typeDefs, err := parseSchema(gqlSDL)
	if err != nil {
		return nil, err
	}
	generator := docGenerator{types: typeDefs}
	for _, doc := range docsList.Docs {
		docs, err := generator.GenerateDocs(doc, docsList.ColName)
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
		docs, err := generator.GenerateDocs(doc, docsList.ColName)
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

func createDocJSON(typeDef *client.CollectionDefinition, doc map[string]any) string {
	sb := strings.Builder{}
	for _, field := range typeDef.Schema.Fields {
		fieldName := field.Name
		if field.IsPrimaryRelation() {
			fieldName += request.RelatedObjectID
		}
		if _, hasProp := doc[fieldName]; !hasProp {
			continue
		}
		format := `"%s": %v`
		if _, isStr := doc[fieldName].(string); isStr {
			format = `"%s": "%v"`
		}
		if sb.Len() == 0 {
			sb.WriteString("{\n")
		} else {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(format, fieldName, doc[fieldName]))
	}
	sb.WriteString("\n}")
	return sb.String()
}

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

func (this *docGenerator) generatePrimary(
	doc map[string]any,
	typeDef *client.CollectionDefinition,
) (map[string]any, []GeneratedDoc, error) {
	result := []GeneratedDoc{}
	requested := toRequestedDoc(doc, typeDef)
	for _, field := range typeDef.Schema.Fields {
		if field.IsRelation() {
			if _, hasProp := doc[field.Name]; hasProp {
				if field.IsPrimaryRelation() {
					subType := this.types[field.Schema]
					subDoc := toRequestedDoc(doc[field.Name].(map[string]any), &subType)
					jsonSubDoc := createDocJSON(&subType, subDoc)
					clientSubDoc, err := client.NewDocFromJSON([]byte(jsonSubDoc))
					if err != nil {
						return nil, nil, NewErrFailedToGenerateDoc(err)
					}
					requested[field.Name+request.RelatedObjectID] = clientSubDoc.Key().String()
					result = append(result, GeneratedDoc{ColName: subType.Description.Name, JSON: jsonSubDoc})
				}
			}
		}
	}
	return requested, result, nil
}

func (this *docGenerator) GenerateDocs(doc map[string]any, typeName string) ([]GeneratedDoc, error) {
	typeDef := this.types[typeName]

	requested, result, err := this.generatePrimary(doc, &typeDef)
	if err != nil {
		return nil, err
	}
	docStr := createDocJSON(&typeDef, requested)

	result = append(result, GeneratedDoc{ColName: typeDef.Description.Name, JSON: docStr})

	var docKey string
	for _, field := range typeDef.Schema.Fields {
		if field.IsRelation() {
			if _, hasProp := doc[field.Name]; hasProp {
				if !field.IsPrimaryRelation() {
					if docKey == "" {
						clientDoc, err := client.NewDocFromJSON([]byte(docStr))
						if err != nil {
							return nil, NewErrFailedToGenerateDoc(err)
						}
						docKey = clientDoc.Key().String()
					}
					docs, err := this.generateSecondaryDocs(doc, typeName, &field, docKey)
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
					actions, err := this.GenerateDocs(relDoc, relTypeDef.Description.Name)
					if err != nil {
						return nil, err
					}
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocKey
				actions, err := this.GenerateDocs(relVal, relTypeDef.Description.Name)
				if err != nil {
					return nil, err
				}
				result = append(result, actions...)
			}
		}
	}
	return result, nil
}
