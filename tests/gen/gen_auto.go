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

const (
	// DefaultNumDocs is the default number of documents to generate for a collection.
	DefaultNumDocs = 10
	// DefaultNumChildrenPerDoc is the default number of children to generate for a document.
	DefaultNumChildrenPerDoc = 2

	// DefaultStrLen is the default length of a string to generate.
	DefaultStrLen = 10
	// DefaultIntMin is the default minimum value of an integer to generate.
	DefaultIntMin = 0
	// DefaultIntMax is the default maximum value of an integer to generate.
	DefaultIntMax = 10000
)

type docRec struct {
	doc    map[string]any
	docKey string
}

// AutoGenerateFromSDL generates random documents from a schema.
func AutoGenerateFromSDL(gqlSDL string, options ...Option) ([]GeneratedDoc, error) {
	genConfigs, err := parseConfig(gqlSDL)
	if err != nil {
		return nil, err
	}
	typeDefs, err := parseSchema(gqlSDL)
	if err != nil {
		return nil, err
	}
	generator := randomDocGenerator{types: typeDefs, config: genConfigs}
	return generator.GenerateDocs(options...)
}

// AutoGenerate generates random documents from collection definitions.
func AutoGenerate(definitions []client.CollectionDefinition, options ...Option) ([]GeneratedDoc, error) {
	err := validateDefinitions(definitions)
	if err != nil {
		return nil, err
	}
	typeDefs := make(map[string]client.CollectionDefinition)
	for _, def := range definitions {
		typeDefs[def.Description.Name] = def
	}
	generator := randomDocGenerator{types: typeDefs}
	return generator.GenerateDocs(options...)
}

type randomDocGenerator struct {
	types        map[string]client.CollectionDefinition
	config       configsMap
	resultDocs   []GeneratedDoc
	usageCounter typeUsageCounters
	cols         map[string][]docRec
	docsDemand   map[string]typeDemand
	random       rand.Rand
}

func (g *randomDocGenerator) GenerateDocs(options ...Option) ([]GeneratedDoc, error) {
	g.cols = make(map[string][]docRec)
	if g.config == nil {
		g.config = make(configsMap)
	}

	configurator := newDocGenConfigurator(g.types, g.config)
	err := configurator.Configure(options...)
	if err != nil {
		return nil, err
	}

	g.random = *configurator.random

	g.docsDemand = configurator.docsDemand
	g.usageCounter = configurator.usageCounter

	g.resultDocs = make([]GeneratedDoc, 0, g.getMaxTotalDemand())
	docsLists, err := g.generateRandomDocs(configurator.typesOrder)
	if err != nil {
		return nil, err
	}
	for _, docsList := range docsLists {
		typeDef := g.types[docsList.ColName]
		for _, doc := range docsList.Docs {
			g.resultDocs = append(g.resultDocs, GeneratedDoc{
				ColName: typeDef.Description.Name,
				JSON:    createDocJSON(&typeDef, doc),
			})
		}
	}
	return g.resultDocs, nil
}

func (g *randomDocGenerator) getMaxTotalDemand() int {
	totalDemand := 0
	for _, demand := range g.docsDemand {
		totalDemand += demand.max
	}
	return totalDemand
}

// getNextPrimaryDocKey returns the key of the next primary document to be used as a relation.
func (g *randomDocGenerator) getNextPrimaryDocKey(secondaryType string, field *client.FieldDescription) string {
	ind := g.usageCounter.getNextTypeIndForField(secondaryType, field)
	docKey := g.cols[field.Schema][ind].docKey
	return docKey
}

func (g *randomDocGenerator) getDocKey(typeDef *client.CollectionDefinition, doc map[string]any) (string, error) {
	clientDoc, err := client.NewDocFromJSON([]byte(createDocJSON(typeDef, doc)))
	if err != nil {
		return "", err
	}
	return clientDoc.Key().String(), nil
}

func (g *randomDocGenerator) generateRandomDocs(order []string) ([]DocsList, error) {
	result := []DocsList{}
	for _, typeName := range order {
		col := DocsList{ColName: typeName}
		typeDef := g.types[typeName]

		currentTypeDemand := g.docsDemand[typeName]
		averageDemand := currentTypeDemand.getAverage()
		for i := 0; i < averageDemand; i++ {
			newDoc := make(map[string]any)
			for _, field := range typeDef.Schema.Fields {
				if field.Name == request.KeyFieldName {
					continue
				}
				if field.IsRelation() {
					if field.IsPrimaryRelation() {
						newDoc[field.Name+request.RelatedObjectID] = g.getNextPrimaryDocKey(typeName, &field)
					}
				} else {
					fieldConf := g.config.ForField(typeName, field.Name)
					newDoc[field.Name] = g.generateRandomValue(typeName, field.Kind, fieldConf)
				}
			}
			docKey, err := g.getDocKey(&typeDef, newDoc)
			if err != nil {
				return nil, err
			}
			g.cols[typeName] = append(g.cols[typeName], docRec{doc: newDoc, docKey: docKey})
			col.Docs = append(col.Docs, newDoc)
		}
		result = append(result, col)
	}
	return result, nil
}

func getRandomString(random *rand.Rand, n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[random.Intn(len(letterBytes))]
	}
	return string(b)
}

func (g *randomDocGenerator) generateRandomValue(
	typeName string,
	fieldKind client.FieldKind,
	fieldConfig genConfig,
) any {
	genVal := g.getValueGenerator(fieldKind, fieldConfig)
	if fieldConfig.fieldGenerator != nil {
		return fieldConfig.fieldGenerator(len(g.cols[typeName]), genVal)
	}
	return genVal()
}

func (g *randomDocGenerator) getValueGenerator(fieldKind client.FieldKind, fieldConfig genConfig) func() any {
	switch fieldKind {
	case client.FieldKind_STRING:
		strLen := DefaultStrLen
		if prop, ok := fieldConfig.props["len"]; ok {
			strLen = prop.(int)
		}
		return func() any { return getRandomString(&g.random, strLen) }
	case client.FieldKind_INT:
		min, max := getMinMaxOrDefault(fieldConfig, DefaultIntMin, DefaultIntMax)
		return func() any { return min + g.random.Intn(max-min+1) }
	case client.FieldKind_BOOL:
		ratio := 0.5
		if prop, ok := fieldConfig.props["ratio"]; ok {
			ratio = prop.(float64)
		}
		return func() any { return g.random.Float64() < ratio }
	case client.FieldKind_FLOAT:
		min, max := getMinMaxOrDefault(fieldConfig, 0.0, 1.0)
		return func() any { return min + g.random.Float64()*(max-min) }
	}
	panic("Can not generate random value for unknown type: " + fieldKind.String())
}

func validateDefinitions(definitions []client.CollectionDefinition) error {
	colIDs := make(map[uint32]struct{})
	colNames := make(map[string]struct{})
	fieldRefs := []string{}
	for _, def := range definitions {
		if def.Description.Name == "" {
			return NewErrIncompleteColDefinition("description name is empty")
		}
		if def.Schema.Name == "" {
			return NewErrIncompleteColDefinition("schema name is empty")
		}
		if def.Description.Name != def.Schema.Name {
			return NewErrIncompleteColDefinition("description name and schema name do not match")
		}
		for _, field := range def.Schema.Fields {
			if field.Name == "" {
				return NewErrIncompleteColDefinition("field name is empty")
			}
			if field.IsObject() {
				if field.Schema == "" {
					return NewErrIncompleteColDefinition("field schema is empty")
				}
				fieldRefs = append(fieldRefs, field.Schema)
			}
		}
		colNames[def.Description.Name] = struct{}{}
		colIDs[def.Description.ID] = struct{}{}
	}
	for _, ref := range fieldRefs {
		if _, ok := colNames[ref]; !ok {
			return NewErrIncompleteColDefinition("field schema references unknown collection")
		}
	}
	if len(colIDs) != len(definitions) {
		return NewErrIncompleteColDefinition("duplicate collection IDs")
	}
	return nil
}
