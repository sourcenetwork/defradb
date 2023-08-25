// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("http")

type logEntry struct {
	req *http.Request
}

var _ middleware.LogEntry = (*logEntry)(nil)

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	log.Info(
		e.req.Context(),
		"Request",
		logging.NewKV("Method", e.req.Method),
		logging.NewKV("Path", e.req.URL.Path),
		logging.NewKV("Status", status),
		logging.NewKV("LengthBytes", bytes),
		logging.NewKV("ElapsedTime", elapsed.String()),
	)
}

func (e *logEntry) Panic(v any, stack []byte) {
	middleware.PrintPrettyStack(v)
}

type logFormatter struct{}

var _ middleware.LogFormatter = (*logFormatter)(nil)

func (f *logFormatter) NewLogEntry(req *http.Request) middleware.LogEntry {
	return &logEntry{req}
}
