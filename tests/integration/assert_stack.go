// Copyright 2024 Democratized Data Foundation
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
	"strconv"
	"strings"
)

// assertStack keeps track of the current assertion path.
// GraphQL response can be traversed by a key of a map and/or an index of an array.
// So whenever we have a mismatch in a large response, we can use this stack to find the exact path.
// Example output: "commits[2].links[1].cid"
type assertStack struct {
	stack []string
	isMap []bool
}

func (a *assertStack) pushMap(key string) {
	a.stack = append(a.stack, key)
	a.isMap = append(a.isMap, true)
}

func (a *assertStack) pushArray(index int) {
	a.stack = append(a.stack, strconv.Itoa(index))
	a.isMap = append(a.isMap, false)
}

func (a *assertStack) pop() {
	a.stack = a.stack[:len(a.stack)-1]
	a.isMap = a.isMap[:len(a.isMap)-1]
}

func (a *assertStack) String() string {
	var b strings.Builder
	for i, key := range a.stack {
		if a.isMap[i] {
			if i > 0 {
				b.WriteString(".")
			}
			b.WriteString(key)
		} else {
			b.WriteString("[")
			b.WriteString(key)
			b.WriteString("]")
		}
	}
	return b.String()
}
