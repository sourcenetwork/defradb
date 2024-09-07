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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
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
	types           map[string]client.CollectionDefinition
	definitionCache client.DefinitionCache

	config       configsMap
	primaryGraph map[string][]string
	typesOrder   []string
	docsDemand   map[string]typeDemand
	usageCounter typeUsageCounters
	random       *rand.Rand
}

type collectionID = uint32

// typeUsageCounters is a map of primary type to secondary type to field name to
// relation usage. This is used to keep track of the usage of each relation.
// Each foreign field has a tracker that keeps track of which and how many of primary
// documents have been used for that foreign field. This is used to ensure that the
// number of documents generated for each primary type is within the range of the
// demand for that type and to guarantee a uniform distribution of the documents.
type typeUsageCounters struct {
	m      map[collectionID]map[string]map[string]*relationUsage
	random *rand.Rand
}

func newTypeUsageCounter(random *rand.Rand) typeUsageCounters {
	return typeUsageCounters{
		m:      make(map[collectionID]map[string]map[string]*relationUsage),
		random: random,
	}
}

// addRelationUsage adds a relation usage tracker for a foreign field.
func (c *typeUsageCounters) addRelationUsage(
	secondaryType string,
	field client.FieldDefinition,
	minPerDoc, maxPerDoc, numDocs int,
) {
	var collectionRoot uint32
	switch kind := field.Kind.(type) {
	case *client.CollectionKind:
		collectionRoot = kind.Root

	default:
		return
	}

	if _, ok := c.m[collectionRoot]; !ok {
		c.m[collectionRoot] = make(map[string]map[string]*relationUsage)
	}
	if _, ok := c.m[collectionRoot][secondaryType]; !ok {
		c.m[collectionRoot][secondaryType] = make(map[string]*relationUsage)
	}
	if _, ok := c.m[collectionRoot][secondaryType][field.Name]; !ok {
		c.m[collectionRoot][secondaryType][field.Name] = newRelationUsage(minPerDoc, maxPerDoc, numDocs, c.random)
	}
}

// getNextTypeIndForField returns the next index to be used for a foreign field.
func (c *typeUsageCounters) getNextTypeIndForField(secondaryType string, field *client.FieldDefinition) int {
	var collectionRoot uint32
	switch kind := field.Kind.(type) {
	case *client.CollectionKind:
		collectionRoot = kind.Root
	}

	current := c.m[collectionRoot][secondaryType][field.Name]
	return current.useNextDocIDIndex()
}

type relationUsage struct {
	// counter is the number of primary documents that have been used for the relation.
	counter int
	// minSecDocsPerPrimary is the minimum number of primary documents that should be used for the relation.
	minSecDocsPerPrimary int
	// maxSecDocsPerPrimary is the maximum number of primary documents that should be used for the relation.
	maxSecDocsPerPrimary int
	// docIDsCounter is a slice of structs that keep track of the number of times
	// each primary document has been used for the relation.
	docIDsCounter []struct {
		// ind is the index of the primary document.
		ind int
		// count is the number of times the primary document has been used for the relation.
		count int
	}
	// numAvailablePrimaryDocs is the number of documents of the primary type that are available
	// for the relation.
	numAvailablePrimaryDocs int
	random                  *rand.Rand
}

func newRelationUsage(minSecDocPerPrim, maxSecDocPerPrim, numDocs int, random *rand.Rand) *relationUsage {
	return &relationUsage{
		minSecDocsPerPrimary:    minSecDocPerPrim,
		maxSecDocsPerPrimary:    maxSecDocPerPrim,
		numAvailablePrimaryDocs: numDocs,
		random:                  random,
	}
}

// useNextDocIDIndex determines the next primary document to be used for the relation, tracks
// it and returns its index.
func (u *relationUsage) useNextDocIDIndex() int {
	docIDCounterInd := 0
	// if a primary document has a minimum number of secondary documents that should be
	// generated for it, then it should be used until that minimum is reached.
	// After that, we can pick a random primary document to use.
	if u.counter >= u.minSecDocsPerPrimary*u.numAvailablePrimaryDocs {
		docIDCounterInd = u.random.Intn(len(u.docIDsCounter))
	} else {
		docIDCounterInd = u.counter % len(u.docIDsCounter)
	}
	currentInd := u.docIDsCounter[docIDCounterInd].ind
	docCounter := &u.docIDsCounter[docIDCounterInd]
	docCounter.count++
	// if the primary document reached max number of secondary documents, we can remove it
	// from the slice of primary documents that are available for the relation.
	if docCounter.count >= u.maxSecDocsPerPrimary {
		lastCounterInd := len(u.docIDsCounter) - 1
		*docCounter = u.docIDsCounter[lastCounterInd]
		u.docIDsCounter = u.docIDsCounter[:lastCounterInd]
	}
	u.counter++

	return currentInd
}

// allocateIndexes allocates the indexes for the relation usage tracker.
func (u *relationUsage) allocateIndexes() {
	docIDsCounter := make([]struct {
		ind   int
		count int
	}, u.numAvailablePrimaryDocs)
	for i := range docIDsCounter {
		docIDsCounter[i].ind = i
	}
	u.docIDsCounter = docIDsCounter
}

func newDocGenConfigurator(types map[string]client.CollectionDefinition, config configsMap) docsGenConfigurator {
	defs := make([]client.CollectionDefinition, 0, len(types))
	for _, def := range types {
		defs = append(defs, def)
	}

	return docsGenConfigurator{
		types:           types,
		definitionCache: client.NewDefinitionCache(defs),
		config:          config,
		docsDemand:      make(map[string]typeDemand),
	}
}

func (g *docsGenConfigurator) Configure(options ...Option) error {
	for _, option := range options {
		option(g)
	}

	for typeName := range g.docsDemand {
		if _, ok := g.types[typeName]; !ok {
			return newNotDefinedTypeErr(typeName)
		}
	}

	err := validateConfig(g.types, g.config)
	if err != nil {
		return err
	}

	if g.random == nil {
		g.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	g.usageCounter = newTypeUsageCounter(g.random)

	g.primaryGraph = g.getRelationGraph(g.types)
	g.typesOrder = getTopologicalOrder(g.primaryGraph, g.types)

	if len(g.docsDemand) == 0 {
		g.docsDemand[g.typesOrder[0]] = typeDemand{min: DefaultNumDocs, max: DefaultNumDocs}
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
	for _, typeName := range g.typesOrder {
		if demand, ok := initialTypes[typeName]; ok {
			var err error
			// from the current type we go up the graph and calculate the demand for primary types
			demand, err = g.getPrimaryDemand(typeName, demand, g.primaryGraph)
			if err != nil {
				return err
			}
			g.docsDemand[typeName] = demand

			err = g.calculateDemandForSecondaryTypes(typeName, g.primaryGraph)
			if err != nil {
				return err
			}
		}
	}

	// for other types that are not in the same graph as the initial types, we start with primary
	// types, give them default demand value and calculate the demand for secondary types.
	for _, typeName := range g.typesOrder {
		if _, ok := g.docsDemand[typeName]; !ok {
			g.docsDemand[typeName] = typeDemand{min: DefaultNumDocs, max: DefaultNumDocs}
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

		def := g.types[typeName]

		for _, usage := range g.usageCounter.m[def.Description.RootID] {
			for _, field := range usage {
				if field.numAvailablePrimaryDocs == math.MaxInt {
					field.numAvailablePrimaryDocs = max
				}
				if field.numAvailablePrimaryDocs > demand.max {
					field.numAvailablePrimaryDocs = demand.max
				}
				field.allocateIndexes()
			}
		}
	}
}

func (g *docsGenConfigurator) getDemandForPrimaryType(
	primaryType, secondaryType string,
	secondaryDemand typeDemand,
	primaryGraph map[string][]string,
) (typeDemand, error) {
	primaryTypeDef := g.types[primaryType]
	secondaryTypeDef := g.types[secondaryType]

	for _, field := range primaryTypeDef.GetFields() {
		var otherRoot immutable.Option[uint32]
		switch kind := field.Kind.(type) {
		case *client.CollectionKind:
			otherRoot = immutable.Some(kind.Root)
		}

		if otherRoot.HasValue() && otherRoot.Value() == secondaryTypeDef.Description.RootID {
			primaryDemand := typeDemand{min: secondaryDemand.min, max: secondaryDemand.max}
			minPerDoc, maxPerDoc := 1, 1

			if field.Kind.IsArray() {
				fieldConf := g.config.ForField(primaryType, field.Name)
				minPerDoc, maxPerDoc = getMinMaxOrDefault(fieldConf, 0, secondaryDemand.max)
				// if we request min 100 of secondary docs and there can be max 5 per primary doc,
				// then we need to generate at least 20 primary docs.
				minRatio := float64(secondaryDemand.min) / float64(maxPerDoc)
				primaryDemand.min = int(math.Ceil(minRatio))
				if minPerDoc == 0 {
					primaryDemand.max = math.MaxInt
				} else {
					// if we request max 200 of secondary docs and there can be min 10 per primary doc,
					// then we need to generate at most 2000 primary docs.
					maxRatio := float64(secondaryDemand.max) / float64(minPerDoc)
					primaryDemand.max = int(math.Floor(maxRatio))
				}

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
				return typeDemand{}, NewErrCanNotSupplyTypeDemand(primaryType)
			}
			g.docsDemand[primaryType] = primaryDemand
			g.initRelationUsages(secondaryTypeDef.GetName(), primaryType, minPerDoc, maxPerDoc)
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
	for _, field := range typeDef.GetFields() {
		if field.Kind.IsObject() && !field.IsPrimaryRelation {
			primaryDocDemand := g.docsDemand[typeName]
			newSecDemand := typeDemand{min: primaryDocDemand.min, max: primaryDocDemand.max}
			minPerDoc, maxPerDoc := 1, 1

			otherType, _ := client.GetDefinition(g.definitionCache, typeDef, field.Kind)
			curSecDemand, hasSecDemand := g.docsDemand[otherType.GetName()]

			if field.Kind.IsArray() {
				fieldConf := g.config.ForField(typeName, field.Name)
				if prop, ok := fieldConf.props["min"]; ok {
					minPerDoc = prop.(int)
					maxPerDoc = fieldConf.props["max"].(int)
					newSecDemand.min = primaryDocDemand.max * minPerDoc
					newSecDemand.max = primaryDocDemand.min * maxPerDoc
				} else if hasSecDemand {
					minPerDoc = curSecDemand.min / primaryDocDemand.max
					maxPerDoc = curSecDemand.max / primaryDocDemand.min
					newSecDemand.min = curSecDemand.min
					newSecDemand.max = curSecDemand.max
				} else {
					minPerDoc = DefaultNumChildrenPerDoc
					maxPerDoc = DefaultNumChildrenPerDoc
					newSecDemand.min = primaryDocDemand.max * minPerDoc
					newSecDemand.max = primaryDocDemand.min * maxPerDoc
				}
			}

			if hasSecDemand {
				if curSecDemand.min < newSecDemand.min || curSecDemand.max > newSecDemand.max {
					return NewErrCanNotSupplyTypeDemand(otherType.GetName())
				}
			} else {
				g.docsDemand[otherType.GetName()] = newSecDemand
			}
			g.initRelationUsages(otherType.GetName(), typeName, minPerDoc, maxPerDoc)

			err := g.calculateDemandForSecondaryTypes(otherType.GetName(), primaryGraph)
			if err != nil {
				return err
			}

			for _, primaryTypeName := range primaryGraph[otherType.GetName()] {
				if _, ok := g.docsDemand[primaryTypeName]; !ok {
					primaryDemand, err := g.getDemandForPrimaryType(
						primaryTypeName,
						otherType.GetName(),
						newSecDemand,
						primaryGraph,
					)
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

func (g *docsGenConfigurator) initRelationUsages(secondaryType, primaryType string, minPerDoc, maxPerDoc int) {
	secondaryTypeDef := g.types[secondaryType]
	primaryTypeDef := g.types[primaryType]
	for _, secondaryTypeField := range secondaryTypeDef.GetFields() {
		var otherRoot immutable.Option[uint32]
		switch kind := secondaryTypeField.Kind.(type) {
		case *client.CollectionKind:
			otherRoot = immutable.Some(kind.Root)
		}

		if otherRoot.HasValue() && otherRoot.Value() == primaryTypeDef.Description.RootID {
			g.usageCounter.addRelationUsage(secondaryType, secondaryTypeField, minPerDoc,
				maxPerDoc, g.docsDemand[primaryType].getAverage())
		}
	}
}

func (g *docsGenConfigurator) getRelationGraph(types map[string]client.CollectionDefinition) map[string][]string {
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
		for _, field := range typeDef.GetFields() {
			if field.Kind.IsObject() {
				otherDef, _ := client.GetDefinition(g.definitionCache, typeDef, field.Kind)

				if field.IsPrimaryRelation {
					primaryGraph[typeName] = appendUnique(primaryGraph[typeName], otherDef.GetName())
				} else {
					primaryGraph[otherDef.GetName()] = appendUnique(primaryGraph[otherDef.GetName()], typeName)
				}
			}
		}
	}

	return primaryGraph
}

func getTopologicalOrder(graph map[string][]string, types map[string]client.CollectionDefinition) []string {
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
