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
	"math/rand"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

func AutoGenerateDocs(schema string, count int) []GeneratedDoc {
	parser := schemaParser{}
	typeDefs, genConfigs := parser.Parse(schema)
	generator := randomDocGenerator{types: typeDefs, config: genConfigs}
	return generator.GenerateDocs(count)
}

type randomDocGenerator struct {
	types      map[typeNameStr]typeDefinition
	config     map[typeNameStr]map[fieldNameStr]genConfig
	resultDocs []GeneratedDoc
	counter    map[typeNameStr]map[fieldNameStr]map[string]int
	cols       map[typeNameStr][]docRec
}

func (g *randomDocGenerator) GenerateDocs(count int) []GeneratedDoc {
	g.resultDocs = make([]GeneratedDoc, 0, count)
	order := findDependencyOrder(g.types)
	docsLists := g.generateRandomDocs(count, order)
	for _, docsList := range docsLists {
		typeDef := g.types[docsList.ColName]
		for _, doc := range docsList.Docs {
			g.resultDocs = append(g.resultDocs, GeneratedDoc{
				ColIndex: typeDef.index,
				JSON:     createDocJSON(doc, &typeDef),
			})
		}
	}
	return g.resultDocs
}

func getRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (g *randomDocGenerator) generateRandomValue(typeStr string, config genConfig) any {
	switch typeStr {
	case "String":
		strLen := 10
		if prop, ok := config.props["len"]; ok {
			strLen = prop.(int)
		}
		return getRandomString(strLen)
	case "Int":
		minInt := 0
		intRange := 10000
		if prop, ok := config.props["min"]; ok {
			minInt = prop.(int)
		}
		if prop, ok := config.props["max"]; ok {
			intRange = prop.(int) - minInt
		}
		return minInt + rand.Intn(intRange)
	case "Boolean":
		return rand.Float32() < 0.5
	case "Float":
		minFloat := 0.0
		floatRange := 1.0
		if prop, ok := config.props["min"]; ok {
			minFloat = prop.(float64)
		}
		if prop, ok := config.props["max"]; ok {
			floatRange = prop.(float64) - minFloat
		}
		return minFloat + rand.Float64()*floatRange
	}
	panic("Can not generate random value for unknown type: " + typeStr)
}

type doc = map[string]any
type docRec struct {
	doc    doc
	docKey string
}

func (g *randomDocGenerator) incrementCounter(secondaryType, secondaryProp, primaryType string) int {
	if g.counter[primaryType] == nil {
		g.counter[primaryType] = make(map[string]map[string]int)
	}
	if g.counter[primaryType][secondaryType] == nil {
		g.counter[primaryType][secondaryType] = make(map[string]int)
	}
	prev := g.counter[primaryType][secondaryType][secondaryProp]
	prev = prev % len(g.cols[primaryType])
	g.counter[primaryType][secondaryType][secondaryProp]++
	return prev
}

func (g *randomDocGenerator) getDocKey(typeName string, ind int) string {
	if g.cols[typeName][ind].docKey == "" {
		typeDef := g.types[typeName]
		clientDoc, err := client.NewDocFromJSON([]byte(createDocJSON(g.cols[typeName][ind].doc, &typeDef)))
		if err != nil {
			panic("Failed to create doc from JSON: " + err.Error())
		}
		g.cols[typeName][ind].docKey = clientDoc.Key().String()
	}
	return g.cols[typeName][ind].docKey
}

func (g *randomDocGenerator) getPrimaryDocKeyForField(typeName string, field fieldDefinition) string {
	relDocInd := g.incrementCounter(typeName, field.name, field.typeStr)
	return g.getDocKey(field.typeStr, relDocInd)
}

func (g *randomDocGenerator) generateRandomDocs(count int, order []string) []DocsList {
	g.counter = make(map[typeNameStr]map[fieldNameStr]map[string]int)
	g.cols = make(map[typeNameStr][]docRec)
	result := []DocsList{}
	docDemands := map[typeNameStr]int{order[0]: count}
	for _, typeName := range order {
		col := DocsList{ColName: typeName}
		docDemand := docDemands[typeName]
		for i := 0; i < docDemand; i++ {
			typeDef := g.types[typeName]
			newDoc := make(doc)
			for _, field := range typeDef.fields {
				if field.isRelation {
					if field.isPrimary {
						newDoc[field.name+request.RelatedObjectID] = g.getPrimaryDocKeyForField(typeName, field)
					} else {
						inc := 1
						if field.isArray {
							inc = 2
						}
						docDemands[field.typeStr] += inc
					}
				} else {
					var fieldConfig genConfig
					typeConfig := g.config[typeName]
					if typeConfig != nil {
						fieldConfig = typeConfig[field.name]
					}
					newDoc[field.name] = g.generateRandomValue(field.typeStr, fieldConfig)
				}
			}
			g.cols[typeName] = append(g.cols[typeName], docRec{doc: newDoc})
			col.Docs = append(col.Docs, newDoc)
		}
		result = append(result, col)
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
		for _, field := range typeDef.fields {
			if field.isRelation {
				if field.isPrimary {
					graph[field.typeStr] = appendUnique(graph[field.typeStr], typeName)
				} else {
					graph[typeName] = appendUnique(graph[typeName], field.typeStr)
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
