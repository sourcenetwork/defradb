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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// createSchemaWithDocs returns UpdateSchema action and CreateDoc actions
// with the documents that match the schema.
// The schema is parsed to get the list of fields, and the docs
// are created with the same fields.
// This allows us to have only one large list of docs with predefined
// fields, and create schemas with different fields from it.
func CreateSchemaWithDocs(schema string, docsList DocsList) []any {
	docs := GenerateDocs(schema, docsList)
	resultActions := make([]any, 0, len(docs)+1)
	resultActions = append(resultActions, testUtils.SchemaUpdate{Schema: schema})
	for _, doc := range docs {
		resultActions = append(resultActions, testUtils.CreateDoc{CollectionID: doc.ColIndex, Doc: doc.JSON})
	}
	return resultActions
}

func GenerateDocs(schema string, docsList DocsList) []GeneratedDoc {
	resultDocs := make([]GeneratedDoc, 0, len(docsList.Docs))
	parser := schemaParser{}
	typeDefs, _ := parser.Parse(schema)
	generator := docGenerator{types: typeDefs}
	for _, doc := range docsList.Docs {
		docs := generator.GenerateDocs(doc, docsList.ColName)
		resultDocs = append(resultDocs, docs...)
	}
	return resultDocs
}

type docGenerator struct {
	types map[string]typeDefinition
}

func createDocJSON(doc map[string]any, typeDef *typeDefinition) string {
	sb := strings.Builder{}
	for fieldName := range doc {
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
) (map[string]any, []GeneratedDoc) {
	result := []GeneratedDoc{}
	requested := toRequestedDoc(doc, typeDef)
	for _, field := range typeDef.fields {
		if field.isRelation {
			if _, hasProp := doc[field.name]; hasProp {
				if field.isPrimary {
					subType := this.types[field.typeStr]
					subDoc := toRequestedDoc(doc[field.name].(map[string]any), &subType)
					jsonSubDoc := createDocJSON(subDoc, &subType)
					clientSubDoc, err := client.NewDocFromJSON([]byte(jsonSubDoc))
					if err != nil {
						panic("Failed to create doc from JSON: " + err.Error())
					}
					requested[field.name+request.RelatedObjectID] = clientSubDoc.Key().String()
					result = append(result, GeneratedDoc{ColIndex: subType.index, JSON: jsonSubDoc})
				}
			}
		}
	}
	return requested, result
}

func (this *docGenerator) GenerateDocs(doc map[string]any, typeName string) []GeneratedDoc {
	typeDef := this.types[typeName]

	requested, result := this.generatePrimary(doc, &typeDef)
	docStr := createDocJSON(requested, &typeDef)

	result = append(result, GeneratedDoc{ColIndex: typeDef.index, JSON: docStr})

	var docKey string
	for _, field := range typeDef.fields {
		if field.isRelation {
			if _, hasProp := doc[field.name]; hasProp {
				if !field.isPrimary {
					if docKey == "" {
						clientDoc, err := client.NewDocFromJSON([]byte(docStr))
						if err != nil {
							panic("Failed to create doc from JSON: " + err.Error())
						}
						docKey = clientDoc.Key().String()
					}
					docs := this.generateSecondaryDocs(doc, typeName, &field, docKey)
					result = append(result, docs...)
				}
			}
		}
	}
	return result
}

func (this *docGenerator) generateSecondaryDocs(
	primaryDoc map[string]any,
	primaryTypeName string,
	relProp *fieldDefinition,
	primaryDocKey string,
) []GeneratedDoc {
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
					actions := this.GenerateDocs(relDoc, relTypeDef.name)
					result = append(result, actions...)
				}
			case map[string]any:
				relVal[primaryPropName] = primaryDocKey
				actions := this.GenerateDocs(relVal, relTypeDef.name)
				result = append(result, actions...)
			}
		}
	}
	return result
}
