// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"strings"
)

// Response:
//     query @explain {
//         user {
//             _key
//         }
//     }

// Input to ExplainAllText:
// [ "selectTopNode" "selectNode" "scanNode" ]

// Outputs:
// ----------------
// Explain Request:
// ----------------
// selectTopNode
//         selectNode
//                 scanNode

// ExplainAllText formats with proper indenting all the node names.
func ExplainAllText(nodesToExplain []string) []map[string]interface{} {

	var explainResponse []map[string]interface{}

	var explainBuilder string = "----------------\nExplain Request:\n----------------\n"

	for i, node := range nodesToExplain {
		explainBuilder = explainBuilder + strings.Repeat("\t", i) + node + "\n"
	}

	// fmt.Println(explainBuilder)

	return append(
		explainResponse,
		map[string]interface{}{
			"explain": explainBuilder,
		},
	)

}
