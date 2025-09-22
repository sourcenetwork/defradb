// Copyright 2025 Democratized Data Foundation
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
	"bytes"
	"html/template"
	"maps"
	"strconv"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/tests/state"
)

// templateDataGenerators contains a set of data generators by their template prefix.
//
// Supporting action properties will replace any templated elements with data drawn from these
// sets.
var templateDataGenerators = map[string]func(*state.State, int) map[string]string{
	"CID": func(s *state.State, nodeID int) map[string]string {

		docIDsToCIDs := s.Nodes[nodeID].Composites

		res := map[string]string{}
		for colIndex, docIndexes := range s.DocIDs {
			for docIndex, docID := range docIndexes {
				cids := docIDsToCIDs[docID.String()]
				for cidIndex, cid := range cids {
					templateCIDRef := "CID" +
						// The index of the collection in the test.
						strconv.Itoa(colIndex) + "_" +
						// The index of the document within that collection.
						strconv.Itoa(docIndex) + "_" +
						// The index of the CID for that document.
						// WARNING: This mights be difficult for the writer of the
						// test to accurately determine when testing P2P functionalities.
						strconv.Itoa(cidIndex)
					res[templateCIDRef] = cid.String()
				}
			}
		}
		return res
	},
}

// replace returns a new string with any templating placholders (see "text/template") with data drawn
// from `state`.
func replace(s *state.State, nodeId int, input string) string {
	if !strings.Contains(input, "{{") {
		// If the input doesn't contain any templating elements we can return early
		return input
	}

	templateData := map[string]string{}
	for _, datasetGenerator := range templateDataGenerators {
		// Having to regenerate the full dataset for every node-action is horribly inefficient, but
		// it is tolerable for now.
		maps.Copy(templateData, datasetGenerator(s, nodeId))
	}

	tmpl := template.Must(template.New("").Parse(input))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, templateData)
	if err != nil {
		require.Fail(s.T, errors.WithStack(err).Error())
	}

	return buf.String()
}
