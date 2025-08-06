// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"bytes"
	"maps"
	"strconv"
	"strings"
	"text/template"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/tests/state"
)

// templateDataGenerators contains a set of data generators by their template prefix.
//
// Supporting action properties will replace any templated elements with data drawn from these
// sets.
var templateDataGenerators = map[string]func(*state.State, int) map[string]string{
	"Policy": func(s *state.State, nodeID int) map[string]string {
		nodesPolicyIDs := s.PolicyIDs[nodeID]

		res := map[string]string{}
		for i, policyID := range nodesPolicyIDs {
			res["Policy"+strconv.Itoa(i)] = policyID
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
