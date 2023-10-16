// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/sourcenetwork/immutable"

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
func createSchemaWithDocs(schema string) []any {
	userDocs := getUserDocs()
	resultActions := make([]any, 0, len(userDocs.docs)+1)
	resultActions = append(resultActions, testUtils.SchemaUpdate{Schema: schema})
	parser := schemaParser{}
	typeDefs := parser.Parse(schema)
	generator := createDocGenerator{types: typeDefs}
	order := findDependencyOrder(typeDefs)
	randomDocsCols := generateRandomDocs(6, typeDefs, order)
	for _, col := range randomDocsCols {
		for _, doc := range col.docs {
			actions := generator.GenerateDocs(doc, col.colName)
			resultActions = append(resultActions, actions...)
		}
	}
	/*for _, doc := range userDocs.docs {
		actions := generator.GenerateDocs(doc, userDocs.colName)
		resultActions = append(resultActions, actions...)
	}*/
	return resultActions
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

func generateRandomDocs(count int, types map[string]typeDefinition, order []string) []docsCollection {
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

	result := []docsCollection{}
	for _, typeName := range order {
		col := docsCollection{colName: typeName}
		for i := 0; i < count; i++ {
			typeDef := types[typeName]
			newDoc := make(doc)
			for _, prop := range typeDef.props {
				if prop.isRelation {
					if prop.isPrimary.Value() {
						relDocInd := incrementCounter(prop.typeStr, typeName, prop.name)
						docKey := getDocKey(prop.typeStr, relDocInd)
						newDoc[prop.name+request.RelatedObjectID] = docKey
					}
				} else {
					newDoc[prop.name] = generateRandomValue(prop.typeStr)
				}
			}
			cols[typeName] = append(cols[typeName], docRec{doc: newDoc})
			col.docs = append(col.docs, newDoc)
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
) (map[string]any, []any) {
	result := []any{}
	requested := toRequestedDoc(doc, typeDef)
	for _, prop := range typeDef.props {
		if prop.isRelation {
			if _, hasProp := doc[prop.name]; hasProp {
				if prop.isPrimary.Value() {
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

func (this *createDocGenerator) GenerateDocs(doc map[string]any, typeName string) []any {
	typeDef := this.types[typeName]

	requested, result := this.generatePrimary(doc, &typeDef)
	docStr := createDocJSON(requested, &typeDef)

	result = append(result, testUtils.CreateDoc{CollectionID: typeDef.index, Doc: docStr})

	var docKey string
	for _, prop := range typeDef.props {
		if prop.isRelation {
			if _, hasProp := doc[prop.name]; hasProp {
				if !prop.isPrimary.Value() {
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
) []any {
	result := []any{}
	relTypeDef := this.types[relProp.typeStr]
	primaryPropName := ""
	for _, relDocProp := range relTypeDef.props {
		if relDocProp.typeStr == primaryTypeName && relDocProp.isPrimary.Value() {
			primaryPropName = relDocProp.name + request.RelatedObjectID
			switch relVal := primaryDoc[relProp.name].(type) {
			case docsCollection:
				for _, relDoc := range relVal.docs {
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

type propDefinition struct {
	name       string
	typeStr    string
	isArray    bool
	isRelation bool
	isPrimary  immutable.Option[bool]
}

type typeDefinition struct {
	name  string
	index int
	props map[string]propDefinition
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
				if propDef.isPrimary.Value() {
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

type schemaParser struct {
	types             map[string]typeDefinition
	schemaLines       []string
	firstRelationType string
	currentTypeDef    typeDefinition
	relationTypesMap  map[string]map[string]string
}

func (p *schemaParser) Parse(schema string) map[string]typeDefinition {
	p.types = make(map[string]typeDefinition)
	p.relationTypesMap = make(map[string]map[string]string)
	p.schemaLines = strings.Split(schema, "\n")
	p.findTypes()

	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.currentTypeDef = p.types[typeName]
			continue
		}
		if strings.HasPrefix(line, "}") {
			p.types[p.currentTypeDef.name] = p.currentTypeDef
			continue
		}
		pos := strings.Index(line, ":")
		if pos != -1 {
			p.defineProp(line, pos)
		}
	}
	p.resolvePrimaryRelations()
	return p.types
}

func (p *schemaParser) findTypes() {
	typeIndex := 0
	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.types[typeName] = typeDefinition{name: typeName, index: typeIndex, props: make(map[string]propDefinition)}
			typeIndex++
		}
	}
}

func (p *schemaParser) defineProp(line string, pos int) {
	prop := propDefinition{name: line[:pos]}
	prop.typeStr = strings.TrimSpace(line[pos+1:])
	typeEndPos := strings.Index(prop.typeStr, " ")
	if typeEndPos != -1 {
		prop.typeStr = prop.typeStr[:typeEndPos]
	}
	if prop.typeStr[0] == '[' {
		prop.isArray = true
		prop.typeStr = prop.typeStr[1 : len(prop.typeStr)-1]
	}
	if _, isRelation := p.types[prop.typeStr]; isRelation {
		prop.isRelation = true
		if prop.isArray {
			prop.isPrimary = immutable.Some(false)
		} else if strings.Contains(line[pos+len(prop.typeStr)+2:], "@primary") {
			prop.isPrimary = immutable.Some(true)
		}
		relMap := p.relationTypesMap[prop.typeStr]
		if relMap == nil {
			relMap = make(map[string]string)
		}
		relMap[prop.name] = p.currentTypeDef.name
		p.relationTypesMap[prop.typeStr] = relMap
		if p.firstRelationType == "" {
			p.firstRelationType = p.currentTypeDef.name
		}
	}
	p.currentTypeDef.props[prop.name] = prop
}

func (p *schemaParser) resolvePrimaryRelations() {
	for typeName, relationProps := range p.relationTypesMap {
		typeDef := p.types[typeName]
		for _, prop := range typeDef.props {
			for relPropName, relPropType := range relationProps {
				if prop.typeStr == relPropType {
					relatedTypeDef := p.types[relPropType]
					relatedProp := relatedTypeDef.props[relPropName]
					if !relatedProp.isPrimary.HasValue() {
						relatedProp.isPrimary = immutable.Some(typeName == p.firstRelationType)
						relatedTypeDef.props[relPropName] = relatedProp
						p.types[relPropType] = relatedTypeDef
						delete(p.relationTypesMap, relPropType)
					}
					if !prop.isPrimary.HasValue() {
						val := typeName != p.firstRelationType
						if relatedProp.isPrimary.HasValue() {
							val = !relatedProp.isPrimary.Value()
						}
						prop.isPrimary = immutable.Some(val)
						typeDef.props[prop.name] = prop
					}
				}
			}
		}
		p.types[typeName] = typeDef
	}
}

func makeExplainQuery(req string) string {
	return "query @explain(type: execute) " + req[6:]
}
