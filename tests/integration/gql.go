// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

// jsonToGQL transforms a json doc string to a gql string.
func jsonToGQL(val string) (string, error) {
	bytes := []byte(val)

	if client.IsJSONArray(bytes) {
		var doc []map[string]any
		if err := json.Unmarshal(bytes, &doc); err != nil {
			return "", err
		}
		return arrayToGQL(doc)
	} else {
		var doc map[string]any
		if err := json.Unmarshal(bytes, &doc); err != nil {
			return "", err
		}
		return mapToGQL(doc)
	}
}

// valueToGQL transforms a value to a gql string.
func valueToGQL(val any) (string, error) {
	switch t := val.(type) {
	case map[string]any:
		return mapToGQL(t)

	case []any:
		return sliceToGQL(t)
	}
	out, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// mapToGQL transforms a map to a gql string.
func mapToGQL(val map[string]any) (string, error) {
	var entries []string
	for k, v := range val {
		out, err := valueToGQL(v)
		if err != nil {
			return "", err
		}
		entries = append(entries, fmt.Sprintf("%s: %s", k, out))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ",")), nil
}

// sliceToGQL transforms a slice to a gql string.
func sliceToGQL(val []any) (string, error) {
	var entries []string
	for _, v := range val {
		out, err := valueToGQL(v)
		if err != nil {
			return "", err
		}
		entries = append(entries, out)
	}
	return fmt.Sprintf("[%s]", strings.Join(entries, ",")), nil
}

// arrayToGQL transforms an array of maps to a gql string.
func arrayToGQL(val []map[string]any) (string, error) {
	var entries []string
	for _, v := range val {
		out, err := mapToGQL(v)
		if err != nil {
			return "", err
		}
		entries = append(entries, out)
	}
	return fmt.Sprintf("[%s]", strings.Join(entries, ",")), nil
}
