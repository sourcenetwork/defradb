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

import "math/rand"

// Option is a function that configures a document generator.
type Option func(*docsGenConfigurator)

// WithTypeDemand configures the demand for a type.
func WithTypeDemand(typeName string, demand int) Option {
	return func(g *docsGenConfigurator) {
		g.docsDemand[typeName] = typeDemand{min: demand, max: demand, usedDefined: true}
	}
}

// WithTypeDemandRange configures the demand range for a type.
func WithTypeDemandRange(typeName string, min, max int) Option {
	return func(g *docsGenConfigurator) {
		g.docsDemand[typeName] = typeDemand{min: min, max: max, usedDefined: true}
	}
}

// WithTypeDemandRange configures the value range for a field.
func WithFieldRange[T int | float64](typeName, fieldName string, min, max T) Option {
	return func(g *docsGenConfigurator) {
		conf := g.config.ForField(typeName, fieldName)
		conf.props["min"] = min
		conf.props["max"] = max
		g.config.AddForField(typeName, fieldName, conf)
	}
}

// WithFieldLen configures the length of a string field.
func WithFieldLen(typeName, fieldName string, length int) Option {
	return func(g *docsGenConfigurator) {
		conf := g.config.ForField(typeName, fieldName)
		conf.props["len"] = length
		g.config.AddForField(typeName, fieldName, conf)
	}
}

// WithFieldLabels configures a custom field value generator.
func WithFieldGenerator(typeName, fieldName string, genFunc GenerateFieldFunc) Option {
	return func(g *docsGenConfigurator) {
		g.config.AddForField(typeName, fieldName, genConfig{fieldGenerator: genFunc})
	}
}

// WithRandomSeed configures the random seed for the document generator.
func WithRandomSeed(seed int64) Option {
	return func(g *docsGenConfigurator) {
		g.random = rand.New(rand.NewSource(seed))
	}
}

// GenerateFieldFunc is a function that provides custom field values
// It is used as an option to the document generator.
// The function receives the index of the document being generated and a function that
// generates the next value in the sequence of values for the field.
type GenerateFieldFunc func(i int, next func() any) any
