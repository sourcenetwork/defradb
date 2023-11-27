// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package predefined

import (
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/gen"
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

func assertDoc(expected map[string]any, actual gen.GeneratedDoc) string {
	actualMap, err := actual.Doc.ToMap()
	if err != nil {
		return "can not convert doc to map: " + err.Error()
	}
	if !areMapsEquivalent(expected, actualMap) {
		return "docs are not equal"
	}
	return ""
}

// assertDocs asserts that the expected docs are equal to the actual docs ignoring order
func assertDocs(expected []map[string]any, actual []gen.GeneratedDoc) string {
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

func mustGetDocKeyFromDocMap(docMap map[string]any) string {
	doc, err := client.NewDocFromMap(docMap)
	if err != nil {
		panic("can not get doc from map" + err.Error())
	}
	return doc.Key().String()
}

func mustAddKeyToDoc(doc map[string]any) map[string]any {
	doc[request.KeyFieldName] = mustGetDocKeyFromDocMap(doc)
	return doc
}

func mustAddKeysToDocs(docs []map[string]any) []map[string]any {
	for i := range docs {
		mustAddKeyToDoc(docs[i])
	}
	return docs
}
