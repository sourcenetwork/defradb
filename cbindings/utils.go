// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import (
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/defradb/acp/identity"
)

type GoCResult struct {
	Status int
	Error  string
	Value  string
}

type GoCOptions struct {
	Version      string
	CollectionID string
	Name         string
	Identity     string
	GetInactive  int
}

type GoNodeInitOptions struct {
	DbPath                   string
	ListeningAddresses       string
	ReplicatorRetryIntervals string
	Peers                    string
	Identity                 identity.Identity
	InMemory                 int
	DisableP2P               int
	DisableAPI               int
	MaxTransactionRetries    int
	EnableNodeACP            int
}

// returnGoC is a helper function that wraps a status, error, and value into a return object
func returnGoC(status int, errortext string, valuetext string) GoCResult {
	return GoCResult{
		Status: status,
		Error:  errortext,
		Value:  valuetext,
	}
}

// marshalJSONToGoCResult is a helper function that marshals an interface into a return object
func marshalJSONToGoCResult(value any) GoCResult {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", string(dataJSON))
}

// splitCommaSeparatedString is a helper function that turns a single string into an array
func splitCommaSeparatedString(baseStr string) []string {
	var retArr []string
	if baseStr != "" {
		retArr = strings.Split(baseStr, ",")
	} else {
		retArr = []string{}
	}
	return retArr
}
