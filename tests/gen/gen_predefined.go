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

// GenerateDocs generates documents for a schema from a predefined list of docs that might
// include nested docs.
// The schema is parsed to get the list of fields, and the docs
// are created with the fields parsed from the schema.
// This allows us to have only one large list of docs with predefined
// fields, and create schemas with different fields from it.
func GenerateDocs(schema string, docsList DocsList) ([]GeneratedDoc, error) {
	resultDocs := make([]GeneratedDoc, 0, len(docsList.Docs))
	parser := schemaParser{}
	typeDefs, _, _ := parser.Parse(schema)
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
	types map[string]typeDefinition
}

func createDocJSON(typeDef *typeDefinition, doc map[string]any) string {
	sb := strings.Builder{}
	for _, field := range typeDef.fields {
		fieldName := field.name
		if field.isRelation && field.isPrimary {
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

func toRequestedDoc(doc map[string]any, typeDef *typeDefinition) map[string]any {
	result := make(map[string]any)
	for _, field := range typeDef.fields {
		if field.isRelation {
			continue
		}
		result[field.name] = doc[field.name]
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
	typeDef *typeDefinition,
) (map[string]any, []GeneratedDoc, error) {
	result := []GeneratedDoc{}
	requested := toRequestedDoc(doc, typeDef)
	for _, field := range typeDef.fields {
		if field.isRelation {
			if _, hasProp := doc[field.name]; hasProp {
				if field.isPrimary {
					subType := this.types[field.typeStr]
					subDoc := toRequestedDoc(doc[field.name].(map[string]any), &subType)
					jsonSubDoc := createDocJSON(&subType, subDoc)
					clientSubDoc, err := client.NewDocFromJSON([]byte(jsonSubDoc))
					if err != nil {
						return nil, nil, NewErrFailedToGenerateDoc(err)
					}
					requested[field.name+request.RelatedObjectID] = clientSubDoc.Key().String()
					result = append(result, GeneratedDoc{ColIndex: subType.index, JSON: jsonSubDoc})
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

	result = append(result, GeneratedDoc{ColIndex: typeDef.index, JSON: docStr})

	var docKey string
	for _, field := range typeDef.fields {
		if field.isRelation {
			if _, hasProp := doc[field.name]; hasProp {
				if !field.isPrimary {
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
	relProp *fieldDefinition,
	primaryDocKey string,
) ([]GeneratedDoc, error) {
	result := []GeneratedDoc{}
	relTypeDef := this.types[relProp.typeStr]
	primaryPropName := ""
	for _, relDocProp := range relTypeDef.fields {
		if relDocProp.typeStr == primaryTypeName && relDocProp.isPrimary {
			primaryPropName = relDocProp.name + request.RelatedObjectID
			switch relVal := primaryDoc[relProp.name].(type) {
			case []map[string]any:
				for _, relDoc := range relVal {
					relDoc[primaryPropName] = primaryDocKey
					actions, err := this.GenerateDocs(relDoc, relTypeDef.name)
					if err != nil {
						return nil, err
					}
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocKey
				actions, err := this.GenerateDocs(relVal, relTypeDef.name)
				if err != nil {
					return nil, err
				}
				result = append(result, actions...)
			}
		}
	}
	return result, nil
}
