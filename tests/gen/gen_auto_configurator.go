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

type typeDemand struct {
	min, max int
}

func (d typeDemand) getAverage() int {
	if d.max == math.MaxInt {
		return d.max
	}
	return (d.min + d.max) / 2
}

type docsGenConfigurator struct {
	types                        map[string]typeDefinition
	config                       configsMap
	primaryGraph, secondaryGraph map[string][]string
	TypesOrder                   []string
	DocsDemand                   map[string]typeDemand
	UsageCounter                 typeUsageCounters
}

type typeUsageCounters struct {
	m map[string]map[string]map[string]*relationUsage
}

func (c typeUsageCounters) addRelationUsage(secondaryType string, field fieldDefinition, min, max, numDocs int) {
	primaryType := field.typeStr
	if _, ok := c.m[primaryType]; !ok {
		c.m[primaryType] = make(map[string]map[string]*relationUsage)
	}
	if _, ok := c.m[primaryType][secondaryType]; !ok {
		c.m[primaryType][secondaryType] = make(map[string]*relationUsage)
	}
	if _, ok := c.m[primaryType][secondaryType][field.name]; !ok {
		c.m[primaryType][secondaryType][field.name] = newRelationUsage(min, max, numDocs)
	}
}

func (c typeUsageCounters) getNextTypeIndForField(secondaryType string, field fieldDefinition) int {
	primaryType := field.typeStr
	current := c.m[primaryType][secondaryType][field.name]

	ind := current.useNextDocKey()
	return ind
}

func (c typeUsageCounters) allocateIndexes(currentMax int) {
	for _, secondaryTypes := range c.m {
		for _, fields := range secondaryTypes {
			for _, field := range fields {
				if field.numDocs == math.MaxInt {
					field.numDocs = currentMax
				}
				field.allocateIndexes()
			}
		}
	}
}

type relationUsage struct {
	index          int
	minAmount      int
	maxAmount      int
	docKeysCounter []struct {
		ind   int
		count int
	}
	numDocs int
}

func newRelationUsage(minAmount, maxAmount, numDocs int) *relationUsage {
	return &relationUsage{
		minAmount: minAmount,
		maxAmount: maxAmount,
		numDocs:   numDocs,
	}
}

func (u *relationUsage) useNextDocKey() int {
	docKeyCounterInd := 0
	if u.index >= u.minAmount*u.numDocs {
		docKeyCounterInd = rand.Intn(len(u.docKeysCounter))
	} else {
		docKeyCounterInd = u.index % len(u.docKeysCounter)
	}
	currentInd := u.docKeysCounter[docKeyCounterInd].ind
	counter := &u.docKeysCounter[docKeyCounterInd]
	counter.count++
	if counter.count >= u.maxAmount {
		lastCounterInd := len(u.docKeysCounter) - 1
		*counter = u.docKeysCounter[lastCounterInd]
		u.docKeysCounter = u.docKeysCounter[:lastCounterInd]
	}
	u.index++

	return currentInd
}

func (u *relationUsage) allocateIndexes() {
	docKeysCounter := make([]struct {
		ind   int
		count int
	}, u.numDocs)
	for i := range docKeysCounter {
		docKeysCounter[i].ind = i
	}
	u.docKeysCounter = docKeysCounter
}

func newDocGenConfigurator(types map[string]typeDefinition, config configsMap) *docsGenConfigurator {
	return &docsGenConfigurator{
		types:        types,
		config:       config,
		DocsDemand:   make(map[string]typeDemand),
		UsageCounter: typeUsageCounters{m: make(map[string]map[string]map[string]*relationUsage)},
	}
}

func (g *docsGenConfigurator) Configure(options ...Option) error {
	for _, option := range options {
		option(g)
	}

	err := validateConfig(g.types, g.config)
	if err != nil {
		return err
	}

	g.primaryGraph, g.secondaryGraph = getRelationGraphs(g.types)
	g.TypesOrder = getTopologicalOrder(g.primaryGraph, g.types)

	if len(g.DocsDemand) == 0 {
		g.DocsDemand[g.TypesOrder[0]] = typeDemand{min: defaultNumDocs, max: defaultNumDocs}
	}

	initialTypes := make(map[string]typeDemand)
	for typeName, typeDemand := range g.DocsDemand {
		initialTypes[typeName] = typeDemand
	}

	for typeName, typeDemand := range initialTypes {
		var err error
		typeDemand, err = g.getPrimaryDemand(typeName, typeDemand, g.primaryGraph)
		if err != nil {
			return err
		}
		g.DocsDemand[typeName] = typeDemand
		err = g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
		if err != nil {
			return err
		}
	}

	for _, typeName := range g.TypesOrder {
		if _, ok := g.DocsDemand[typeName]; !ok {
			g.DocsDemand[typeName] = typeDemand{min: defaultNumDocs, max: defaultNumDocs}
			err := g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
			if err != nil {
				return err
			}
		}
	}
	g.allocateUsageCounterIndexes()
	return nil
}

func (g *docsGenConfigurator) allocateUsageCounterIndexes() {
	max := 0
	for _, demand := range g.DocsDemand {
		if demand.max > max && demand.max != math.MaxInt {
			max = demand.max
		}
	}
	for typeName, demand := range g.DocsDemand {
		if demand.max == math.MaxInt {
			demand.max = max
			demand.min = max
			g.DocsDemand[typeName] = demand
		}
	}
	g.UsageCounter.allocateIndexes(max)
}

func (g *docsGenConfigurator) getDemandForPrimaryType(
	primaryType, secondaryType string,
	secondaryDemand typeDemand,
	primaryGraph map[string][]string,
) (typeDemand, error) {
	primaryTypeDef := g.types[primaryType]
	for _, field := range primaryTypeDef.fields {
		if field.isRelation && field.typeStr == secondaryType {
			primaryDemand := secondaryDemand
			min, max := 1, 1
			if field.isArray {
				fieldConf := g.config.ForField(primaryType, field.name)
				min, max = getMinMaxOrDefault(fieldConf, 0, secondaryDemand.max)
				minRatio := float64(secondaryDemand.min) / float64(max)
				maxRatio := float64(secondaryDemand.max) / float64(min)
				primaryDemand.min = int(math.Ceil(minRatio))
				primaryDemand.max = int(math.Floor(maxRatio))
				var err error
				primaryDemand, err = g.getPrimaryDemand(primaryType, primaryDemand, primaryGraph)
				if err != nil {
					return typeDemand{}, err
				}
				if tmp, ok := g.DocsDemand[primaryType]; ok {
					if primaryDemand.min < tmp.min {
						primaryDemand.min = tmp.min
					}
					if primaryDemand.max > tmp.max {
						primaryDemand.max = tmp.max
					}
				}
				if primaryDemand.min > primaryDemand.max {
					return typeDemand{}, NewErrInvalidConfiguration("can not supply demand for type " + primaryType)
				}
			}
			g.DocsDemand[primaryType] = primaryDemand
			g.initRelationUsages(field.typeStr, primaryType, min, max)
		}
	}
	return secondaryDemand, nil
}

func (g *docsGenConfigurator) getPrimaryDemand(
	secondaryType string,
	secondaryDemand typeDemand,
	primaryGraph map[string][]string,
) (typeDemand, error) {
	for _, primaryTypeName := range primaryGraph[secondaryType] {
		var err error
		secondaryDemand, err = g.getDemandForPrimaryType(primaryTypeName, secondaryType, secondaryDemand, primaryGraph)
		if err != nil {
			return typeDemand{}, err
		}
	}
	return secondaryDemand, nil
}

func (g *docsGenConfigurator) calculateDemandForSecondaryTypes(
	typeName string,
	primaryGraph map[string][]string,
) error {
	typeDef := g.types[typeName]
	for _, field := range typeDef.fields {
		if field.isRelation && !field.isPrimary {
			primaryDocDemand := g.DocsDemand[typeName]
			secondaryDocDemand := primaryDocDemand
			min, max := 1, 1

			if field.isArray {
				min, max = getMinMaxOrDefault(g.config.ForField(typeName, field.name), 2, 2)
				secondaryDocDemand.max = primaryDocDemand.min * max
				secondaryDocDemand.min = primaryDocDemand.max * min
			}

			g.DocsDemand[field.typeStr] = secondaryDocDemand
			g.initRelationUsages(field.typeStr, typeName, min, max)
			err := g.calculateDemandForSecondaryTypes(field.typeStr, primaryGraph)
			if err != nil {
				return err
			}

			for _, primaryTypeName := range primaryGraph[field.typeStr] {
				if _, ok := g.DocsDemand[primaryTypeName]; !ok {
					primaryDemand, err := g.getDemandForPrimaryType(primaryTypeName, field.typeStr, secondaryDocDemand, primaryGraph)
					if err != nil {
						return err
					}
					g.DocsDemand[primaryTypeName] = primaryDemand
				}
			}
		}
	}
	return nil
}

func (g *docsGenConfigurator) initRelationUsages(secondaryType, primaryType string, min, max int) {
	secondaryTypeDef := g.types[secondaryType]
	for _, secondaryTypeField := range secondaryTypeDef.fields {
		if secondaryTypeField.typeStr == primaryType {
			g.UsageCounter.addRelationUsage(secondaryType, secondaryTypeField, min, max, g.DocsDemand[primaryType].getAverage())
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
