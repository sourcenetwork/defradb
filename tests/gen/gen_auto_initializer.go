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
)

type docsGenInitializer struct {
	types                        map[tStr]typeDefinition
	config                       configsMap
	primaryGraph, secondaryGraph map[string][]string
	TypesOrder                   []string
	DocsDemand                   map[tStr]int
	UsageCounter                 map[tStr]map[tStr]map[fStr]relationUsage
}

func newDocsDemandCalculator(types map[tStr]typeDefinition, config configsMap) *docsGenInitializer {
	return &docsGenInitializer{
		types:        types,
		config:       config,
		DocsDemand:   make(map[tStr]int),
		UsageCounter: make(map[tStr]map[tStr]map[fStr]relationUsage),
	}
}

func (g *docsGenInitializer) Init(colName string, count int) {
	g.primaryGraph, g.secondaryGraph = getRelationGraphs(g.types)
	g.TypesOrder = getTopologicalOrder(g.primaryGraph, g.types)

	getFirstTypeWithDemand := func() (string, int) {
		for _, typeName := range g.TypesOrder {
			if demand, ok := g.DocsDemand[typeName]; ok {
				return typeName, demand
			}
		}
		panic("No types with demand")
	}

	g.DocsDemand[colName] = count

	typeName, demand := getFirstTypeWithDemand()
	demand = g.getPrimaryDemand(typeName, demand, g.primaryGraph)
	g.DocsDemand[typeName] = demand
	typeName, _ = getFirstTypeWithDemand()
	g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
	for _, typeName := range g.TypesOrder {
		if _, ok := g.DocsDemand[typeName]; !ok {
			g.DocsDemand[typeName] = defaultNumDocs
			g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
		}
	}
}

func (g *docsGenInitializer) getDemandForPrimaryType(
	primaryType, secondaryType string,
	currentDemand int,
	primaryGraph map[string][]string,
) int {
	primaryTypeDef := g.types[primaryType]
	for _, field := range primaryTypeDef.fields {
		if field.isRelation && field.typeStr == secondaryType {
			min, max := 1, 1
			if field.isArray {
				min, max = getMinMaxOrDefault(g.config.ForField(primaryType, field.name),
					defaultNumChildrenPerDoc, defaultNumChildrenPerDoc)
				average := float64(min) + float64(max-min)/2
				currentDemand = int(math.Ceil(float64(currentDemand) / average))
				if currentDemand == 0 {
					currentDemand = 1
				}
				currentDemand = g.getPrimaryDemand(primaryType, currentDemand, primaryGraph)
				tmp := g.DocsDemand[primaryType]
				if tmp > currentDemand {
					currentDemand = tmp
				}
			}
			g.DocsDemand[primaryType] = currentDemand
			g.initRelationUsages(field.typeStr, primaryType, min, max)
		}
	}
	return currentDemand
}

func (g *docsGenInitializer) getPrimaryDemand(
	secondaryType string,
	currentDemand int,
	primaryGraph map[string][]string,
) int {
	for _, primaryTypeName := range primaryGraph[secondaryType] {
		currentDemand = g.getDemandForPrimaryType(primaryTypeName, secondaryType, currentDemand, primaryGraph)
	}
	return currentDemand
}

func (g *docsGenInitializer) calculateDemandForSecondaryTypes(typeName string, primaryGraph map[string][]string) {
	typeDef := g.types[typeName]
	for _, field := range typeDef.fields {
		if field.isRelation && !field.isPrimary {
			primaryDocDemand := g.DocsDemand[typeName]
			secondaryDocDemand := primaryDocDemand
			min, max := 1, 1

			if field.isArray {
				min, max = getMinMaxOrDefault(g.config.ForField(typeName, field.name), 2, 2)
				average := float64(min) + float64(max-min)/2
				secondaryDocDemand = int(float64(primaryDocDemand) * average)
			}

			g.DocsDemand[field.typeStr] = secondaryDocDemand
			g.initRelationUsages(field.typeStr, typeName, min, max)
			g.calculateDemandForSecondaryTypes(field.typeStr, primaryGraph)

			for _, primaryTypeName := range primaryGraph[field.typeStr] {
				if _, ok := g.DocsDemand[primaryTypeName]; !ok {
					g.DocsDemand[primaryTypeName] = g.getDemandForPrimaryType(primaryTypeName, field.typeStr,
						secondaryDocDemand, primaryGraph)
				}
			}
		}
	}
}

func (g *docsGenInitializer) initRelationUsages(secondaryType, primaryType string, min, max int) {
	secondaryTypeDef := g.types[secondaryType]
	for _, secondaryTypeField := range secondaryTypeDef.fields {
		if secondaryTypeField.typeStr == primaryType {
			g.addRelationUsage(secondaryType, secondaryTypeField, min, max)
		}
	}
}

func (g *docsGenInitializer) addRelationUsage(secondaryType string, field fieldDefinition, min, max int) {
	primaryType := field.typeStr
	if _, ok := g.UsageCounter[primaryType]; !ok {
		g.UsageCounter[primaryType] = make(map[tStr]map[fStr]relationUsage)
	}
	if _, ok := g.UsageCounter[primaryType][secondaryType]; !ok {
		g.UsageCounter[primaryType][secondaryType] = make(map[fStr]relationUsage)
	}
	if _, ok := g.UsageCounter[primaryType][secondaryType][field.name]; !ok {
		g.UsageCounter[primaryType][secondaryType][field.name] = newRelationUsage(
			min, max, g.DocsDemand[primaryType])
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
