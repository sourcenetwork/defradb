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
	"math/rand"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// createSchemaWithDocs returns UpdateSchema action and CreateDoc actions
// with the documents that match the schema.
// The schema is parsed to get the list of properties, and the docs
// are created with the same properties.
// This allows us to have only one large list of docs with predefined
// properties, and create schemas with different properties from it.
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
	resultActions := make([]testUtils.CreateDoc, 0, len(docsList.Docs))
	parser := schemaParser{}
	typeDefs := parser.Parse(schema)
	generator := createDocGenerator{types: typeDefs}
	for _, doc := range docsList.Docs {
		actions := generator.GenerateDocs(doc, docsList.ColName)
		resultActions = append(resultActions, actions...)
	}
	res := make([]GeneratedDoc, len(resultActions))
	for i, action := range resultActions {
		res[i] = GeneratedDoc{ColIndex: action.CollectionID, JSON: action.Doc}
	}
	return res
}

func getRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func generateRandomValue(typeStr string) any {
	switch typeStr {
	case "String":
		return getRandomString(10)
	case "Int":
		return rand.Intn(100)
	case "Boolean":
		return rand.Float32() < 0.5
	case "Float":
		return rand.Float64()
	}
	panic("Can not generate random value for unknown type: " + typeStr)
}

type doc = map[string]any
type docRec struct {
	doc    doc
	docKey string
}

func generateRandomDocs(count int, types map[string]typeDefinition, order []string) []DocsList {
	counter := make(map[string]map[string]map[string]int)
	cols := make(map[string][]docRec)
	incrementCounter := func(primary, secondary, secondaryProp string) int {
		if counter[primary] == nil {
			counter[primary] = make(map[string]map[string]int)
		}
		if counter[primary][secondary] == nil {
			counter[primary][secondary] = make(map[string]int)
		}
		prev := counter[primary][secondary][secondaryProp]
		if prev >= len(cols[primary]) {
			panic(fmt.Sprintf("Not enough docs for type %s", primary))
		}
		counter[primary][secondary][secondaryProp]++
		return prev
	}

	getDocKey := func(typeName string, ind int) string {
		if cols[typeName][ind].docKey == "" {
			typeDef := types[typeName]
			clientDoc, err := client.NewDocFromJSON([]byte(createDocJSON(cols[typeName][ind].doc, &typeDef)))
			if err != nil {
				panic("Failed to create doc from JSON: " + err.Error())
			}
			cols[typeName][ind].docKey = clientDoc.Key().String()
		}
		return cols[typeName][ind].docKey
	}

	result := []DocsList{}
	for _, typeName := range order {
		col := DocsList{ColName: typeName}
		for i := 0; i < count; i++ {
			typeDef := types[typeName]
			newDoc := make(doc)
			for _, prop := range typeDef.props {
				if prop.isRelation {
					if prop.isPrimary {
						relDocInd := incrementCounter(prop.typeStr, typeName, prop.name)
						docKey := getDocKey(prop.typeStr, relDocInd)
						newDoc[prop.name+request.RelatedObjectID] = docKey
					}
				} else {
					newDoc[prop.name] = generateRandomValue(prop.typeStr)
				}
			}
			cols[typeName] = append(cols[typeName], docRec{doc: newDoc})
			col.Docs = append(col.Docs, newDoc)
		}
		result = append(result, col)
	}
	return result
}

type createDocGenerator struct {
	types map[string]typeDefinition
}

func createDocJSON(doc map[string]any, typeDef *typeDefinition) string {
	sb := strings.Builder{}
	for propName := range doc {
		format := `"%s": %v`
		if _, isStr := doc[propName].(string); isStr {
			format = `"%s": "%v"`
		}
		if sb.Len() == 0 {
			sb.WriteString("{\n")
		} else {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(format, propName, doc[propName]))
	}
	sb.WriteString("\n}")
	return sb.String()
}

func toRequestedDoc(doc map[string]any, typeDef *typeDefinition) map[string]any {
	result := make(map[string]any)
	for _, prop := range typeDef.props {
		if prop.isRelation {
			continue
		}
		result[prop.name] = doc[prop.name]
	}
	for name, val := range doc {
		if strings.HasSuffix(name, request.RelatedObjectID) {
			result[name] = val
		}
	}
	return result
}

func (this *createDocGenerator) generatePrimary(
	doc map[string]any,
	typeDef *typeDefinition,
) (map[string]any, []testUtils.CreateDoc) {
	result := []testUtils.CreateDoc{}
	requested := toRequestedDoc(doc, typeDef)
	for _, prop := range typeDef.props {
		if prop.isRelation {
			if _, hasProp := doc[prop.name]; hasProp {
				if prop.isPrimary {
					subType := this.types[prop.typeStr]
					subDoc := toRequestedDoc(doc[prop.name].(map[string]any), &subType)
					jsonSubDoc := createDocJSON(subDoc, &subType)
					clientSubDoc, err := client.NewDocFromJSON([]byte(jsonSubDoc))
					if err != nil {
						panic("Failed to create doc from JSON: " + err.Error())
					}
					requested[prop.name+request.RelatedObjectID] = clientSubDoc.Key().String()
					result = append(result, testUtils.CreateDoc{CollectionID: subType.index, Doc: jsonSubDoc})
				}
			}
		}
	}
	return requested, result
}

func (this *createDocGenerator) GenerateDocs(doc map[string]any, typeName string) []testUtils.CreateDoc {
	typeDef := this.types[typeName]

	requested, result := this.generatePrimary(doc, &typeDef)
	docStr := createDocJSON(requested, &typeDef)

	result = append(result, testUtils.CreateDoc{CollectionID: typeDef.index, Doc: docStr})

	var docKey string
	for _, prop := range typeDef.props {
		if prop.isRelation {
			if _, hasProp := doc[prop.name]; hasProp {
				if !prop.isPrimary {
					if docKey == "" {
						clientDoc, err := client.NewDocFromJSON([]byte(docStr))
						if err != nil {
							panic("Failed to create doc from JSON: " + err.Error())
						}
						docKey = clientDoc.Key().String()
					}
					actions := this.generateSecondaryDocs(doc, typeName, &prop, docKey)
					result = append(result, actions...)
				}
			}
		}
	}
	return result
}

func (this *createDocGenerator) generateSecondaryDocs(
	primaryDoc map[string]any,
	primaryTypeName string,
	relProp *propDefinition,
	primaryDocKey string,
) []testUtils.CreateDoc {
	result := []testUtils.CreateDoc{}
	relTypeDef := this.types[relProp.typeStr]
	primaryPropName := ""
	for _, relDocProp := range relTypeDef.props {
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

func findDependencyOrder(parsedTypes map[string]typeDefinition) []string {
	graph := make(map[string][]string)
	visited := make(map[string]bool)
	stack := []string{}

	appendUnique := func(slice []string, val string) []string {
		for _, item := range slice {
			if item == val {
				return slice
			}
		}
		return append(slice, val)
	}

	for typeName, typeDef := range parsedTypes {
		for _, propDef := range typeDef.props {
			if propDef.isRelation {
				if propDef.isPrimary {
					graph[propDef.typeStr] = appendUnique(graph[propDef.typeStr], typeName)
				} else {
					graph[typeName] = appendUnique(graph[typeName], propDef.typeStr)
				}
			}
		}
	}

	var dfs func(node string)
	dfs = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			}
		}

		stack = append(stack, node)
	}

	for typeName := range parsedTypes {
		if !visited[typeName] {
			dfs(typeName)
		}
	}

	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}

	return stack
}
