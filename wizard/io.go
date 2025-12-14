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

// getYAMLValue is a helper function that gets a value from a YAML map
func getYAMLValue(m map[string]any, path []string) any {
	current := m

	for i, key := range path {
		val, exists := current[key]
		if !exists {
			return nil
		}

		// If this is the last element in the path, return its value directly
		if i == len(path)-1 {
			return val
		}

		// Otherwise we expect it to be another nested map
		next, ok := val.(map[string]any)
		if !ok {
			return nil
		}

		current = next
	}

	return nil
}

// setYAMLValue is a helper function that sets a value in a YAML map
func setYAMLValue(m map[string]any, path []string, value any) error {
	current := m
	// Iterate through the path, looking for the deepest element in it
	for i, key := range path {
		// If this is the last element in the path, set the value...
		if i == len(path)-1 {
			current[key] = value
			return nil
		}
		// ...otherwise, proceed to the next element deeper in the map...
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			// ...unless it doesn't exist, in which case create a new map for it
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}
	return nil
}

// setYAMLValueInFile opens a YAML file, updates a value in it, and writes it back
func setYAMLValueInFile(filename string, target []string, value any) error {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal into a map
	var m map[string]any
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

// getYAMLValueInFile opens a YAML file, gets a value from it, and returns the value
func getYAMLValueInFile(filename string, target []string) any {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}

	// Unmarshal into a map
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil
	}

	// Get the value using the helper
	return getYAMLValue(m, target)
}
