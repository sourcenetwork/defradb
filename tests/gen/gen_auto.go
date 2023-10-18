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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

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
