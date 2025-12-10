// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wizard

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// getConfigFile returns the path to $HOME/.defradb/config.yaml
func getConfigFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".defradb", "config.yaml")
}

// setYAMLValue is a helper function that sets a value in a YAML map
func setYAMLValue(m map[string]interface{}, path []string, value interface{}) error {
	current := m
	for i, key := range path {
		if i == len(path)-1 {
			current[key] = value
			return nil
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
	return nil
}

// setYAMLValueInFile opens a YAML file, updates a value in it, and writes it back
func setYAMLValueInFile(filename string, target []string, value interface{}) error {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal into a map
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return err
	}

	// Set the value using the helper
	if err := setYAMLValue(m, target, value); err != nil {
		return err
	}

	// Marshal back to YAML
	newData, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	// Write back to the same file
	return os.WriteFile(filename, newData, 0644)
}
