// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package dac

import (
	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/local"
)

const localStoreName = "local_document_acp"

var _ DocumentACP = (*bridgeDocumentACP)(nil)
var _ acp.ACPSystemClient = (*LocalDocumentACP)(nil)

// LocalDocumentACP represents a local document acp implementation that makes no remote calls.
type LocalDocumentACP struct {
	*local.LocalACP
}

func NewLocalDocumentACP(pathToStore string) (DocumentACP, error) {
	localACP, err := local.NewLocalACP(pathToStore, localStoreName)
	if err != nil {
		return nil, err
	}

	return &bridgeDocumentACP{
		clientACP: &LocalDocumentACP{LocalACP: localACP},
	}, nil
}
