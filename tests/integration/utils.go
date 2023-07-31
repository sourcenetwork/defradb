// Copyright 2022 Democratized Data Foundation
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
	"testing"
)

type RequestTestCase struct {
	Description string
	Request     string

	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	Docs map[int][]string

	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	Updates map[int]map[int][]string

	Results []map[string]any

	// The expected content of an expected error
	ExpectedError string
}

func ExecuteRequestTestCase(
	t *testing.T,
	schema string,
	collectionNames []string,
	test RequestTestCase,
) {
	actions := []any{
		SchemaUpdate{
			Schema: schema,
		},
	}

	for collectionIndex, docs := range test.Docs {
		for _, doc := range docs {
			actions = append(
				actions,
				CreateDoc{
					CollectionID: collectionIndex,
					Doc:          doc,
				},
			)
		}
	}

	for collectionIndex, docUpdates := range test.Updates {
		for docIndex, docs := range docUpdates {
			for _, doc := range docs {
				actions = append(
					actions,
					UpdateDoc{
						CollectionID: collectionIndex,
						DocID:        docIndex,
						Doc:          doc,
					},
				)
			}
		}
	}

	if test.Request != "" {
		actions = append(
			actions,
			Request{
				ExpectedError: test.ExpectedError,
				Request:       test.Request,
				Results:       test.Results,
			},
		)
	}

	ExecuteTestCase(
		t,
		TestCase{
			Description: test.Description,
			Actions:     actions,
		},
	)
}
