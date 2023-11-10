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
	"time"
)

type typeDemand struct {
	min, max    int
	usedDefined bool
}

func (d typeDemand) getAverage() int {
	if d.max == math.MaxInt {
		return d.max
	}
	return (d.min + d.max) / 2
}

// docsGenConfigurator is responsible for handling the provided configuration and
// configuring the document generator. This includes things like setting up the
// demand for each type, setting up the relation usage counters, and setting up
// the random seed.
type docsGenConfigurator struct {
	types        map[string]typeDefinition
	config       configsMap
	primaryGraph map[string][]string
	typesOrder   []string
	docsDemand   map[string]typeDemand
	usageCounter typeUsageCounters
	random       *rand.Rand
}

// typeUsageCounters is a map of primary type to secondary type to field name to
// relation usage. This is used to keep track of the usage of each relation.
// Each foreign field has a tracker that keeps track of which and how many of primary
// documents have been used for that foreign field. This is used to ensure that the
// number of documents generated for each primary type is within the range of the
// demand for that type and to guarantee a uniform distribution of the documents.
type typeUsageCounters struct {
	m      map[string]map[string]map[string]*relationUsage
	random *rand.Rand
}

func newTypeUsageCounter(random *rand.Rand) typeUsageCounters {
	return typeUsageCounters{
		m:      make(map[string]map[string]map[string]*relationUsage),
		random: random,
	}
}

// addRelationUsage adds a relation usage tracker for a foreign field.
func (c *typeUsageCounters) addRelationUsage(secondaryType string, field fieldDefinition, min, max, numDocs int) {
	primaryType := field.typeStr
	if _, ok := c.m[primaryType]; !ok {
		c.m[primaryType] = make(map[string]map[string]*relationUsage)
	}
	if _, ok := c.m[primaryType][secondaryType]; !ok {
		c.m[primaryType][secondaryType] = make(map[string]*relationUsage)
	}
	if _, ok := c.m[primaryType][secondaryType][field.name]; !ok {
		c.m[primaryType][secondaryType][field.name] = newRelationUsage(min, max, numDocs, c.random)
	}
}

// getNextTypeIndForField returns the next index to be used for a foreign field.
func (c *typeUsageCounters) getNextTypeIndForField(secondaryType string, field fieldDefinition) int {
	current := c.m[field.typeStr][secondaryType][field.name]
	return current.useNextDocKey()
}

// allocateIndexes allocates the indexes for each relation usage tracker.
// It is called when all the demand for each type has been calculated.
func (c *typeUsageCounters) allocateIndexes(currentMaxDemand int) {
	for _, secondaryTypes := range c.m {
		for _, fields := range secondaryTypes {
			for _, field := range fields {
				if field.numAvailableDocs == math.MaxInt {
					field.numAvailableDocs = currentMaxDemand
				}
				field.allocateIndexes()
			}
		}
	}
}

type relationUsage struct {
	// counter is the number of primary documents that have been used for the relation.
	counter int
	// minAmount is the minimum number of primary documents that should be used for the relation.
	minAmount int
	// maxAmount is the maximum number of primary documents that should be used for the relation.
	maxAmount int
	// docKeysCounter is a slice of structs that keep track of the number of times
	// each primary document has been used for the relation.
	docKeysCounter []struct {
		// ind is the index of the primary document.
		ind int
		// count is the number of times the primary document has been used for the relation.
		count int
	}
	// numAvailableDocs is the number of documents of the primary type that are available
	// for the relation.
	numAvailableDocs int
	random           *rand.Rand
}

func newRelationUsage(minAmount, maxAmount, numDocs int, random *rand.Rand) *relationUsage {
	return &relationUsage{
		minAmount:        minAmount,
		maxAmount:        maxAmount,
		numAvailableDocs: numDocs,
		random:           random,
	}
}

// useNextDocKey determines the next primary document to be used for the relation, tracks
// it and returns its index.
func (u *relationUsage) useNextDocKey() int {
	docKeyCounterInd := 0
	// if a primary document has a minimum number of secondary documents that should be
	// generated for it, then it should be used until that minimum is reached.
	// After that, we can pick a random primary document to use.
	if u.counter >= u.minAmount*u.numAvailableDocs {
		docKeyCounterInd = u.random.Intn(len(u.docKeysCounter))
	} else {
		docKeyCounterInd = u.counter % len(u.docKeysCounter)
	}
	currentInd := u.docKeysCounter[docKeyCounterInd].ind
	docCounter := &u.docKeysCounter[docKeyCounterInd]
	docCounter.count++
	// if the primary document reached max number of secondary documents, we can remove it
	// from the slice of primary documents that are available for the relation.
	if docCounter.count >= u.maxAmount {
		lastCounterInd := len(u.docKeysCounter) - 1
		*docCounter = u.docKeysCounter[lastCounterInd]
		u.docKeysCounter = u.docKeysCounter[:lastCounterInd]
	}
	u.counter++

	return currentInd
}

func (u *relationUsage) allocateIndexes() {
	docKeysCounter := make([]struct {
		ind   int
		count int
	}, u.numAvailableDocs)
	for i := range docKeysCounter {
		docKeysCounter[i].ind = i
	}
	u.docKeysCounter = docKeysCounter
}

func newDocGenConfigurator(types map[string]typeDefinition, config configsMap) *docsGenConfigurator {
	return &docsGenConfigurator{
		types:      types,
		config:     config,
		docsDemand: make(map[string]typeDemand),
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

	if g.random == nil {
		g.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	g.usageCounter = newTypeUsageCounter(g.random)

	g.primaryGraph = getRelationGraph(g.types)
	g.typesOrder = getTopologicalOrder(g.primaryGraph, g.types)

	if len(g.docsDemand) == 0 {
		g.docsDemand[g.typesOrder[0]] = typeDemand{min: defaultNumDocs, max: defaultNumDocs}
	}

	initialTypes := make(map[string]typeDemand)
	for typeName, typeDemand := range g.docsDemand {
		initialTypes[typeName] = typeDemand
	}

	err = g.calculateDocsDemand(initialTypes)
	if err != nil {
		return err
	}

	g.allocateUsageCounterIndexes()
	return nil
}

func (g *docsGenConfigurator) calculateDocsDemand(initialTypes map[string]typeDemand) error {
	for typeName, typeDemand := range initialTypes {
		var err error
		// from the current type we go up the graph and calculate the demand for primary types
		typeDemand, err = g.getPrimaryDemand(typeName, typeDemand, g.primaryGraph)
		if err != nil {
			return err
		}
		g.docsDemand[typeName] = typeDemand

		err = g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
		if err != nil {
			return err
		}
	}

	// for other types that are not in the same graph as the initial types, we start with primary
	// types, give them default demand value and calculate the demand for secondary types.
	for _, typeName := range g.typesOrder {
		if _, ok := g.docsDemand[typeName]; !ok {
			g.docsDemand[typeName] = typeDemand{min: defaultNumDocs, max: defaultNumDocs}
			err := g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// allocateUsageCounterIndexes allocates the indexes for each relation usage tracker.
func (g *docsGenConfigurator) allocateUsageCounterIndexes() {
	max := 0
	for _, demand := range g.docsDemand {
		if demand.max > max && demand.max != math.MaxInt {
			max = demand.max
		}
	}
	for typeName, demand := range g.docsDemand {
		if demand.max == math.MaxInt {
			demand.max = max
			demand.min = max
			g.docsDemand[typeName] = demand
		}
	}
	g.usageCounter.allocateIndexes(max)
}

func (g *docsGenConfigurator) getDemandForPrimaryType(
	primaryType, secondaryType string,
	secondaryDemand typeDemand,
	primaryGraph map[string][]string,
) (typeDemand, error) {
	primaryTypeDef := g.types[primaryType]
	for _, field := range primaryTypeDef.fields {
		if field.isRelation && field.typeStr == secondaryType {
			primaryDemand := typeDemand{min: secondaryDemand.min, max: secondaryDemand.max}
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
			}
			if currentDemand, ok := g.docsDemand[primaryType]; ok {
				if primaryDemand.min < currentDemand.min {
					primaryDemand.min = currentDemand.min
				}
				if primaryDemand.max > currentDemand.max {
					primaryDemand.max = currentDemand.max
				}
			}

			if primaryDemand.min > primaryDemand.max {
				return typeDemand{}, NewErrInvalidConfiguration("can not supply demand for type " + primaryType)
			}
			g.docsDemand[primaryType] = primaryDemand
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
			primaryDocDemand := g.docsDemand[typeName]
			newSecDemand := typeDemand{min: primaryDocDemand.min, max: primaryDocDemand.max}
			min, max := 1, 1

			if field.isArray {
				fieldConf := g.config.ForField(typeName, field.name)
				min, max = getMinMaxOrDefault(fieldConf, defaultNumChildrenPerDoc, defaultNumChildrenPerDoc)
				newSecDemand.max = primaryDocDemand.min * max
				newSecDemand.min = primaryDocDemand.max * min
			}

			curSecDemand := g.docsDemand[field.typeStr]
			if curSecDemand.usedDefined &&
				(curSecDemand.min < newSecDemand.min || curSecDemand.max > newSecDemand.max) {
				return NewErrInvalidConfiguration("can not supply demand for type " + field.typeStr)
			}
			g.docsDemand[field.typeStr] = newSecDemand
			g.initRelationUsages(field.typeStr, typeName, min, max)

			err := g.calculateDemandForSecondaryTypes(field.typeStr, primaryGraph)
			if err != nil {
				return err
			}

			for _, primaryTypeName := range primaryGraph[field.typeStr] {
				if _, ok := g.docsDemand[primaryTypeName]; !ok {
					primaryDemand, err := g.getDemandForPrimaryType(primaryTypeName, field.typeStr, newSecDemand, primaryGraph)
					if err != nil {
						return err
					}
					g.docsDemand[primaryTypeName] = primaryDemand
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
			g.usageCounter.addRelationUsage(secondaryType, secondaryTypeField, min, max, g.docsDemand[primaryType].getAverage())
		}
	}
}

func getRelationGraph(types map[string]typeDefinition) map[string][]string {
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
					primaryGraph[typeName] = appendUnique(primaryGraph[typeName], field.typeStr)
				} else {
					primaryGraph[field.typeStr] = appendUnique(primaryGraph[field.typeStr], typeName)
				}
			}
		}
	}

	return primaryGraph
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
