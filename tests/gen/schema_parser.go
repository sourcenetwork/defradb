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
	"strconv"
	"strings"
)

type schemaParser struct {
	types            map[string]typeDefinition
	schemaLines      []string
	currentTypeDef   typeDefinition
	relationTypesMap map[string]map[string]string
	resolvedRelation map[string]map[string]bool
}

func (p *schemaParser) Parse(schema string) (map[string]typeDefinition, map[string]map[string]genConfig) {
	p.types = make(map[string]typeDefinition)
	p.relationTypesMap = make(map[string]map[string]string)
	p.resolvedRelation = make(map[string]map[string]bool)
	p.schemaLines = strings.Split(schema, "\n")
	p.findTypes()

	genConfigs := make(map[string]map[string]genConfig)

	var currentConfig map[string]genConfig

	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.currentTypeDef = p.types[typeName]
			currentConfig = make(map[string]genConfig)
			continue
		}
		if strings.HasPrefix(line, "}") {
			p.types[p.currentTypeDef.name] = p.currentTypeDef
			if len(currentConfig) > 0 {
				genConfigs[p.currentTypeDef.name] = currentConfig
			}
			continue
		}
		pos := strings.Index(line, ":")
		if pos != -1 {
			p.defineProp(line, pos)
		}
		configPos := strings.Index(line, "#")
		if configPos != -1 {
			currentConfig[line[:pos]] = p.parseGenConfig(line[configPos+1:])
		}
	}
	p.resolvePrimaryRelations()
	return p.types, genConfigs
}

func (p *schemaParser) parseGenConfig(configStr string) genConfig {
	configStr = strings.TrimSpace(configStr)
	if configStr == "" {
		return genConfig{}
	}

	config := genConfig{props: make(map[string]any)}
	configParts := strings.Split(configStr, ",")
	for _, part := range configParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		propParts := strings.Split(part, ":")
		if len(propParts) == 1 {
			config.labels = append(config.labels, strings.TrimSpace(propParts[0]))
		} else {
			config.props[strings.TrimSpace(propParts[0])] = p.parseGenConfigValue(
				strings.TrimSpace(propParts[1]))
		}
	}
	if len(config.props) == 0 {
		config.props = nil
	}

	return config
}

func (p *schemaParser) parseGenConfigValue(valueStr string) any {
	valueStr = strings.TrimSpace(valueStr)
	if valueStr == "true" {
		return true
	}
	if valueStr == "false" {
		return false
	}
	if valueStr[0] == '"' {
		return valueStr[1 : len(valueStr)-1]
	}
	if strings.Contains(valueStr, ".") {
		if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return val
		}
	}
	if val, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int(val)
	}
	panic(fmt.Sprintf("Failed to parse gen config value %s", valueStr))
}

func (p *schemaParser) findTypes() {
	typeIndex := 0
	for _, line := range p.schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			typeName := strings.TrimSpace(line[5 : 5+typeNameEndPos])
			p.types[typeName] = typeDefinition{name: typeName, index: typeIndex}
			p.resolvedRelation[typeName] = make(map[string]bool)
			typeIndex++
		}
	}
}

func (p *schemaParser) defineProp(line string, pos int) {
	prop := fieldDefinition{name: line[:pos]}
	prop.typeStr = strings.TrimSpace(line[pos+1:])
	typeEndPos := strings.Index(prop.typeStr, " ")
	if typeEndPos != -1 {
		prop.typeStr = prop.typeStr[:typeEndPos]
	}
	if prop.typeStr[0] == '[' {
		prop.isArray = true
		prop.typeStr = prop.typeStr[1 : len(prop.typeStr)-1]
	}
	if _, isRelation := p.types[prop.typeStr]; isRelation {
		prop.isRelation = true
		if prop.isArray {
			prop.isPrimary = false
			p.resolvedRelation[p.currentTypeDef.name][prop.name] = true
		} else if strings.Contains(line[pos+len(prop.typeStr)+2:], "@primary") {
			prop.isPrimary = true
			p.resolvedRelation[p.currentTypeDef.name][prop.name] = true
		}
		relMap := p.relationTypesMap[prop.typeStr]
		if relMap == nil {
			relMap = make(map[string]string)
		}
		relMap[prop.name] = p.currentTypeDef.name
		p.relationTypesMap[prop.typeStr] = relMap
	}
	p.currentTypeDef.fields = append(p.currentTypeDef.fields, prop)
}

func (p *schemaParser) resolvePrimaryField(typeDef, relatedTypeDef *typeDefinition, prop, relatedProp *fieldDefinition) {
	val := typeDef.index < relatedTypeDef.index
	_, isResolved := p.resolvedRelation[typeDef.name][prop.name]
	if isResolved {
		val = !prop.isPrimary
	}
	relatedProp.isPrimary = val
	p.resolvedRelation[relatedTypeDef.name][relatedProp.name] = true
	p.types[relatedTypeDef.name] = *relatedTypeDef
	delete(p.relationTypesMap, prop.typeStr)
}

func (p *schemaParser) resolvePrimaryRelations() {
	for typeName, relationProps := range p.relationTypesMap {
		typeDef := p.types[typeName]
		for i := range typeDef.fields {
			prop := &typeDef.fields[i]
			for relPropName, relPropType := range relationProps {
				if prop.typeStr == relPropType {
					relatedTypeDef := p.types[relPropType]
					relatedProp := relatedTypeDef.getField(relPropName)
					if !p.resolvedRelation[relPropType][relPropName] {
						p.resolvePrimaryField(&typeDef, &relatedTypeDef, prop, relatedProp)
					}
					if !p.resolvedRelation[typeName][prop.name] {
						p.resolvePrimaryField(&relatedTypeDef, &typeDef, relatedProp, prop)
					}
				}
			}
		}
		p.types[typeName] = typeDef
	}
}
