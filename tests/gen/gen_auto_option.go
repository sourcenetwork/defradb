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

type Option func(*docsGenConfigurator)

func WithTypeDemand(typeName string, demand int) Option {
	return func(g *docsGenConfigurator) {
		g.docsDemand[typeName] = typeDemand{min: demand, max: demand}
	}
}

func WithTypeDemandRange(typeName string, min, max int) Option {
	return func(g *docsGenConfigurator) {
		g.docsDemand[typeName] = typeDemand{min: min, max: min}
	}
}

func WithFieldRange[T int | float64](typeName, fieldName string, min, max T) Option {
	return func(g *docsGenConfigurator) {
		conf := g.config.ForField(typeName, fieldName)
		conf.props["min"] = min
		conf.props["max"] = max
		g.config.AddForField(typeName, fieldName, conf)
	}
}

func WithFieldLen(typeName, fieldName string, length int) Option {
	return func(g *docsGenConfigurator) {
		conf := g.config.ForField(typeName, fieldName)
		conf.props["len"] = length
		g.config.AddForField(typeName, fieldName, conf)
	}
}

func WithFieldGenerator(typeName, fieldName string, genFunc GenerateFieldFunc) Option {
	return func(g *docsGenConfigurator) {
		g.config.AddForField(typeName, fieldName, genConfig{fieldGenerator: genFunc})
	}
}
