// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package trigram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{name: "short", value: "ab"},
		{name: "exact", value: "abc", expected: []string{"abc"}},
		{name: "overlapping", value: "abcd", expected: []string{"abc", "bcd"}},
		{name: "lowercase and distinct", value: "AbAbA", expected: []string{"aba", "bab"}},
		{name: "bytes", value: "éa", expected: []string{"éa"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Extract(test.value))
		})
	}
}
