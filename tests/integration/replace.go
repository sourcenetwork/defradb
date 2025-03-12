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

import "github.com/sourcenetwork/defradb/errors"

var (
	_ ReplaceType = (*replacePolicyIndex)(nil)
)

type ReplaceType interface {
	Replacer(input any) (string, error)
}

type replacePolicyIndex struct {
	value int
}

func (r replacePolicyIndex) Replacer(input any) (string, error) {
	// Ensure policy index specified is valid (compared the existing policyIDs) for this node.
	nodesPolicyIDs, ok := input.([]string)
	if !ok {
		return "", errors.New("incorrect policyIDs input")
	}

	policyReplaceIndex := r.value
	if policyReplaceIndex >= len(nodesPolicyIDs) {
		return "", errors.New("a policyID index is out of range, number of added policies is smaller")
	}

	return nodesPolicyIDs[policyReplaceIndex], nil
}

func NewPolicyIndex(value int) replacePolicyIndex {
	return replacePolicyIndex{
		value: value,
	}
}
