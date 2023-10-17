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
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
)

func areValuesEquivalent(a, b any) bool {
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)

	return strA == strB
}

func areMapsEquivalent(m1, m2 map[string]any) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if !areValuesEquivalent(v, m2[k]) {
			return false
		}
	}
	return true
}

func assertDoc(expected map[string]any, actual GeneratedDoc) string {
	actualDocMap := make(map[string]any)
	err := json.Unmarshal([]byte(actual.JSON), &actualDocMap)
	if err != nil {
		return err.Error()
	}
	if !areMapsEquivalent(expected, actualDocMap) {
		return "docs are not equal"
	}
	return ""
}

// assertDocs asserts that the expected docs are equal to the actual docs ignoring order
func assertDocs(expected []map[string]any, actual []GeneratedDoc) string {
	if len(expected) != len(actual) {
		return fmt.Sprintf("expected len %d, got %d", len(expected), len(actual))
	}
outer:
	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(actual); j++ {
			errorMsg := assertDoc(expected[i], actual[j])
			if errorMsg == "" {
				actual = append(actual[:j], actual[j+1:]...)
				continue outer
			}
		}
		return fmt.Sprintf("expected doc not found: %v", expected[i])
	}

	return ""
}

func getDocKey(docMap map[string]any) string {
	docJSON, err := json.Marshal(docMap)
	if err != nil {
		panic("can not marshal doc " + err.Error())
	}
	doc, err := client.NewDocFromJSON([]byte(docJSON))
	if err != nil {
		panic("can not create doc from JSON " + err.Error())
	}
	return doc.Key().String()
}
