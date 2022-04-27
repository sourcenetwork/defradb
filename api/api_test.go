// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package api

import (
	"testing"

	"github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer(nil, nil)

	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(":303000"))
	}

}

func TestServerWithConfig(t *testing.T) {
	s := NewHTTPServer(nil, &http.HandlerConfig{
		Logger: logging.MustNewLogger("defra.http.test"),
	})

	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(":303000"))
	}
}
