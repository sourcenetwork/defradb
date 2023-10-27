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
	"math"
	"math/rand"
)

func (g *randomDocGenerator) calculateDocsDemand(order []tStr, primaryGraph, secondaryGraph map[string][]string) {
	getFirstTypeWithDemand := func() (string, int) {
		for _, typeName := range order {
			if demand, ok := g.docsDemand[typeName]; ok {
				return typeName, demand
			}
		}
		panic("No types with demand")
	}

	typeName, demand := getFirstTypeWithDemand()
	demand = g.getPrimaryDemand(typeName, demand, primaryGraph)
	g.docsDemand[typeName] = demand
	typeName, _ = getFirstTypeWithDemand()
	g.calculateDemandForSecondaryTypes(typeName, primaryGraph)
	for _, typeName := range order {
		if _, ok := g.docsDemand[typeName]; !ok {
			g.docsDemand[typeName] = defaultNumDocs
			g.calculateDemandForSecondaryTypes(typeName, primaryGraph)
		}
	}
}

func (g *randomDocGenerator) getDemandForPrimaryType(
	primaryType, secondaryType string,
	currentDemand int,
	primaryGraph map[string][]string,
) int {
	primaryTypeDef := g.types[primaryType]
	for _, field := range primaryTypeDef.fields {
		if field.isRelation && field.typeStr == secondaryType {
			min, max := 1, 1
			if field.isArray {
				min, max = getMinMaxOrDefault(g.getFieldConfig(primaryType, field.name),
					defaultNumChildrenPerDoc, defaultNumChildrenPerDoc)
				average := float64(min) + float64(max-min)/2
				currentDemand = int(math.Ceil(float64(currentDemand) / average))
				if currentDemand == 0 {
					currentDemand = 1
				}
				currentDemand = g.getPrimaryDemand(primaryType, currentDemand, primaryGraph)
				tmp := g.docsDemand[primaryType]
				if tmp > currentDemand {
					currentDemand = tmp
				}
			}
			g.docsDemand[primaryType] = currentDemand
			g.initRelationUsages(field.typeStr, primaryType, min, max)
		}
	}
	return currentDemand
}

func (g *randomDocGenerator) getPrimaryDemand(
	secondaryType string,
	currentDemand int,
	primaryGraph map[string][]string,
) int {
	for _, primaryTypeName := range primaryGraph[secondaryType] {
		currentDemand = g.getDemandForPrimaryType(primaryTypeName, secondaryType, currentDemand, primaryGraph)
	}
	return currentDemand
}
func getRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (g *randomDocGenerator) calculateDemandForSecondaryTypes(typeName string, primaryGraph map[string][]string) {
	typeDef := g.types[typeName]
	for _, field := range typeDef.fields {
		if field.isRelation && !field.isPrimary {
			primaryDocDemand := g.docsDemand[typeName]
			secondaryDocDemand := primaryDocDemand
			min, max := 1, 1

			if field.isArray {
				min, max = getMinMaxOrDefault(g.getFieldConfig(typeName, field.name), 2, 2)
				average := float64(min) + float64(max-min)/2
				secondaryDocDemand = int(float64(primaryDocDemand) * average)
			}

			g.docsDemand[field.typeStr] = secondaryDocDemand
			g.initRelationUsages(field.typeStr, typeName, min, max)
			g.calculateDemandForSecondaryTypes(field.typeStr, primaryGraph)

			for _, primaryTypeName := range primaryGraph[field.typeStr] {
				if _, ok := g.docsDemand[primaryTypeName]; !ok {
					g.docsDemand[primaryTypeName] = g.getDemandForPrimaryType(primaryTypeName, field.typeStr,
						secondaryDocDemand, primaryGraph)
				}
			}
		}
	}
}

func getRelationGraphs(types map[string]typeDefinition) (map[string][]string, map[string][]string) {
	secondaryGraph := make(map[string][]string)
	primaryGraph := make(map[string][]string)

	appendUnique := func(slice []string, val string) []string {
		for _, item := range slice {
			if item == val {
				return slice
			}
		}
		return append(slice, val)
	}

	for typeName, typeDef := range types {
		for _, field := range typeDef.fields {
			if field.isRelation {
				if field.isPrimary {
					secondaryGraph[field.typeStr] = appendUnique(secondaryGraph[field.typeStr], typeName)
					primaryGraph[typeName] = appendUnique(primaryGraph[typeName], field.typeStr)
				} else {
					secondaryGraph[typeName] = appendUnique(secondaryGraph[typeName], field.typeStr)
					primaryGraph[field.typeStr] = appendUnique(primaryGraph[field.typeStr], typeName)
				}
			}
		}
	}

	return primaryGraph, secondaryGraph
}

func getTopologicalOrder(graph map[string][]string, types map[string]typeDefinition) []string {
	visited := make(map[string]bool)
	stack := []string{}

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

	for typeName := range types {
		if !visited[typeName] {
			dfs(typeName)
		}
	}

	return stack
}
