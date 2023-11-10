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
	"reflect"
)

// genConfig is a configuration for a generation of a field.
type genConfig struct {
	labels         []string
	props          map[string]any
	fieldGenerator GenerateFieldFunc
}

// configsMap is a map of type name to a map of field name to a generation configuration.
type configsMap map[string]map[string]genConfig

// ForField returns the generation configuration for a specific field of a type.
func (m configsMap) ForField(typeStr, fieldName string) genConfig {
	var fieldConfig genConfig
	typeConfig := m[typeStr]
	if typeConfig != nil {
		fieldConfig = typeConfig[fieldName]
	}
	return fieldConfig
}

// AddForField adds a generation configuration for a specific field of a type.
func (m configsMap) AddForField(typeStr, fieldName string, conf genConfig) {
	typeConfig, ok := m[typeStr]
	if !ok {
		typeConfig = make(map[string]genConfig)
		m[typeStr] = typeConfig
	}
	typeConfig[fieldName] = conf
	m[typeStr] = typeConfig
}

func validateConfig(types map[string]typeDefinition, configsMap configsMap) error {
	for typeName, typeConfigs := range configsMap {
		typeDef := types[typeName]
		for fieldName, fieldConfig := range typeConfigs {
			fieldDef := typeDef.getField(fieldName)
			_, hasMin := fieldConfig.props["min"]
			if hasMin {
				var err error
				if fieldDef.isArray || fieldDef.typeStr == intType {
					err = validateMinConfig[int](&fieldConfig, fieldDef.isArray)
				} else {
					err = validateMinConfig[float64](&fieldConfig, false)
				}
				if err != nil {
					return err
				}
			} else if _, hasMax := fieldConfig.props["max"]; hasMax {
				return NewErrInvalidConfiguration("max value is set, but min value is not set")
			}
			lenConf, hasLen := fieldConfig.props["len"]
			if hasLen {
				if fieldDef.typeStr != stringType {
					return NewErrInvalidConfiguration("len val is used on  not String")
				}
				len, ok := lenConf.(int)
				if !ok {
					return NewErrInvalidConfiguration("len value is not integer")
				}
				if len < 1 {
					return NewErrInvalidConfiguration("len value is less than 1")
				}
			}
		}
	}
	return nil
}

func validateMinConfig[T int | float64](fieldConf *genConfig, onlyPositive bool) error {
	min, ok := fieldConf.props["min"].(T)
	if !ok {
		var t T
		return NewErrInvalidConfiguration("min value on array is not " + reflect.TypeOf(t).Name())
	}
	if min < 0 && onlyPositive {
		return NewErrInvalidConfiguration("min value on array is less than 0")
	}
	if maxProp, hasMax := fieldConf.props["max"]; hasMax {
		max, ok := maxProp.(T)
		if !ok && onlyPositive {
			var t T
			return NewErrInvalidConfiguration("max value for array is not " + reflect.TypeOf(t).Name())
		}
		if min > max {
			return NewErrInvalidConfiguration("min value on array is greater than max value")
		}
	} else {
		return NewErrInvalidConfiguration("min value is set, but max value is not set")
	}
	return nil
}

func getMinMaxOrDefault[T int | float64](conf genConfig, min, max T) (T, T) {
	if prop, ok := conf.props["min"]; ok {
		min = prop.(T)
	}
	if prop, ok := conf.props["max"]; ok {
		max = prop.(T)
	}
	return min, max
}
