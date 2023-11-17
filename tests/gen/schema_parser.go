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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/request/graphql"
)

func parseSchema(schema string) (map[string]client.CollectionDefinition, error) {
	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}
	cols, err := parser.ParseSDL(context.Background(), schema)
	if err != nil {
		return nil, err
	}
	result := make(map[string]client.CollectionDefinition)
	for _, col := range cols {
		result[col.Description.Name] = col
	}
	return result, nil
}

func parseConfig(schema string) (configsMap, error) {
	genConfigs := make(map[string]map[string]genConfig)

	var currentConfig map[string]genConfig
	var currentType string

	schemaLines := strings.Split(schema, "\n")
	for _, line := range schemaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") {
			typeNameEndPos := strings.Index(line[5:], " ")
			currentType = strings.TrimSpace(line[5 : 5+typeNameEndPos])
			currentConfig = make(map[string]genConfig)
			continue
		}
		if strings.HasPrefix(line, "}") {
			if len(currentConfig) > 0 {
				genConfigs[currentType] = currentConfig
			}
			continue
		}
		pos := strings.Index(line, ":")
		configPos := strings.Index(line, "#")
		if configPos != -1 {
			var err error
			currentConfig[line[:pos]], err = parseGenConfig(line[configPos+1:])
			if err != nil {
				return nil, err
			}
		}
	}
	return genConfigs, nil
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
