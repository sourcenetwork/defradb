// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package predefined

import (
	"context"
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

func mustGetDocIDFromDocMap(ctx context.Context, docMap map[string]any, collection client.CollectionVersion) string {
	doc, err := client.NewDocFromMap(ctx, docMap, collection)
	if err != nil {
		panic("can not get doc from map" + err.Error())
	}
	return doc.ID().String()
}

func mustAddDocIDToDoc(ctx context.Context, doc map[string]any, collection client.CollectionVersion) map[string]any {
	doc[request.DocIDFieldName] = mustGetDocIDFromDocMap(ctx, doc, collection)
	return doc
}

func mustAddDocIDsToDocs(ctx context.Context, docs []map[string]any, collection client.CollectionVersion) []map[string]any {
	for i := range docs {
		mustAddDocIDToDoc(ctx, docs[i], collection)
	}
	return docs
}
