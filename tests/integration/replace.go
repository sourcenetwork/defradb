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

package tests

import (
	"maps"
	"strconv"
	"strings"

	"github.com/sourcenetwork/defradb/tests/state"
)

// templateDataGenerators contains a set of data generators by their template prefix.
//
// Supporting action properties will replace any templated elements with data drawn from these
// sets.
var templateDataGenerators = map[string]func(*state.State, int) map[string]string{
	"CID": func(s *state.State, nodeID int) map[string]string {
		s.Nodes[nodeID].CompositesLock.RLock()
		defer s.Nodes[nodeID].CompositesLock.RUnlock()
		docIDsToCIDs := s.Nodes[nodeID].Composites

		res := map[string]string{}
		s.DocIDsLock.RLock()
		defer s.DocIDsLock.RUnlock()
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
	"LensID": func(s *state.State, _ int) map[string]string {
		res := map[string]string{}
		for i, lensID := range s.LensIDs {
			templateRef := "LensID" + strconv.Itoa(i)
			res[templateRef] = lensID
		}
		return res
	},
	"CollectionVersionID": func(s *state.State, nodeID int) map[string]string {
		res := map[string]string{}
		for i, versionID := range s.CollectionVersions {
			res["CollectionVersionID"+strconv.Itoa(i)] = versionID
		}
		return res
	},
}

func replaceMap(s *state.State, nodeId int, inputSet []string) map[string]string {
	templateData := map[string]string{}
	for _, datasetGenerator := range templateDataGenerators {
		// Having to regenerate the full dataset for every node-action is horribly inefficient, but
		// it is tolerable for now.
		maps.Copy(templateData, datasetGenerator(s, nodeId))
	}

	result := make(map[string]string, len(inputSet))
	for _, input := range inputSet {
		// WARNING - This does not respect the full Go-replace syntax, at the momement it is a
		// very simple/lightweight key-lookup.  We may want to change this in the future.

		inputID := strings.TrimPrefix(input, "{{.")
		inputID = strings.TrimSuffix(inputID, "}}")

		replacement, ok := templateData[inputID]
		if ok {
			result[input] = replacement
		} else {
			result[input] = input
		}
	}

	return result
}
