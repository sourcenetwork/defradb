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
	"context"
	"strconv"
	"strings"
	"unicode"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/request/graphql"
)

func parseSDL(gqlSDL string) (map[string]client.CollectionDefinition, error) {
	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}
	cols, err := parser.ParseSDL(context.Background(), gqlSDL)
	if err != nil {
		return nil, err
	}
	result := make(map[string]client.CollectionDefinition)
	for _, col := range cols {
		result[col.Description.Name.Value()] = col
	}
	return result, nil
}

func parseConfig(gqlSDL string) (configsMap, error) {
	parser := configParser{}
	err := parser.parse(gqlSDL)
	if err != nil {
		return nil, err
	}
	return parser.genConfigs, nil
}

const typePrefix = "type"

type configParser struct {
	genConfigs     map[string]map[string]genConfig
	currentConfig  map[string]genConfig
	currentType    string
	expectTypeName bool
}

func (p *configParser) tryParseTypeName(line string) {
	const typePrefixLen = len(typePrefix)
	if strings.HasPrefix(line, typePrefix) && (len(line) == typePrefixLen ||
		unicode.IsSpace(rune(line[typePrefixLen]))) {
		p.expectTypeName = true
		line = strings.TrimSpace(line[typePrefixLen:])
	}

	if !p.expectTypeName || line == "" {
		return
	}

	typeNameEndPos := strings.Index(line, " ")
	if typeNameEndPos == -1 {
		typeNameEndPos = len(line)
	}
	p.currentType = strings.TrimSpace(line[:typeNameEndPos])
	p.currentConfig = make(map[string]genConfig)
	p.expectTypeName = false
}

func (p *configParser) tryParseConfig(line string) (bool, error) {
	configPos := strings.Index(line, "#")
	if configPos != -1 {
		var err error
		pos := strings.LastIndex(line[:configPos], ":")
		if pos == -1 {
			return true, nil
		}
		fields := strings.Fields(line[:pos])
		propName := fields[len(fields)-1]
		p.currentConfig[propName], err = parseGenConfig(line[configPos+1:])
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (p *configParser) parseLine(line string) error {
	line = strings.TrimSpace(line)
	if p.currentType == "" {
		p.tryParseTypeName(line)
	}
	skipLine, err := p.tryParseConfig(line)
	if err != nil {
		return err
	}
	if skipLine {
		return nil
	}
	closeTypePos := strings.Index(line, "}")
	if closeTypePos != -1 {
		if len(p.currentConfig) > 0 {
			p.genConfigs[p.currentType] = p.currentConfig
		}
		p.currentType = ""
		return p.parseLine(line[closeTypePos+1:])
	}
	return nil
}

func (p *configParser) parse(gqlSDL string) error {
	p.genConfigs = make(map[string]map[string]genConfig)

	schemaLines := strings.Split(gqlSDL, "\n")
	for _, line := range schemaLines {
		err := p.parseLine(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseGenConfig(configStr string) (genConfig, error) {
	configStr = strings.TrimSpace(configStr)
	if configStr == "" {
		return genConfig{}, nil
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
			if strings.Contains(part, " ") {
				return genConfig{}, NewErrFailedToParse("Config label should not contain spaces: " + configStr)
			}
			config.labels = append(config.labels, strings.TrimSpace(propParts[0]))
		} else {
			propName := strings.TrimSpace(propParts[0])
			if propName == "" {
				return genConfig{}, NewErrFailedToParse("Config property is missing a name: " + configStr)
			}
			propVal := strings.TrimSpace(propParts[1])
			if propVal == "" {
				return genConfig{}, NewErrFailedToParse("Config property is missing a value: " + configStr)
			}
			val, err := parseGenConfigValue(propVal)
			if err != nil {
				return genConfig{}, err
			}
			config.props[propName] = val
		}
	}
	if len(config.props) == 0 {
		config.props = nil
	}

	return config, nil
}

func parseGenConfigValue(valueStr string) (any, error) {
	valueStr = strings.TrimSpace(valueStr)
	if valueStr == "true" {
		return true, nil
	}
	if valueStr == "false" {
		return false, nil
	}
	if valueStr[0] == '"' {
		return valueStr[1 : len(valueStr)-1], nil
	}
	if strings.Contains(valueStr, ".") {
		if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return val, nil
		}
	}
	if val, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int(val), nil
	}
	return nil, NewErrFailedToParse("Failed to parse config value " + valueStr)
}
