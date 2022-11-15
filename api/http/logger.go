// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/logging"
)

type loggingResponseWriter struct {
	statusCode    int
	contentLength int

	http.ResponseWriter
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		statusCode:     http.StatusOK,
		contentLength:  0,
		ResponseWriter: w,
	}
}

func (lrw *loggingResponseWriter) Flush() {
	lrw.ResponseWriter.(http.Flusher).Flush()
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.contentLength = len(b)
	return lrw.ResponseWriter.Write(b)
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(rw)
		next.ServeHTTP(lrw, req)
		elapsed := time.Since(start)
		log.Info(
			req.Context(),
			"Request",
			logging.NewKV(
				"Method",
				req.Method,
			),
			logging.NewKV(
				"Path",
				req.URL.Path,
			),
			logging.NewKV(
				"Status",
				lrw.statusCode,
			),
			logging.NewKV(
				"LengthBytes",
				lrw.contentLength,
			),
			logging.NewKV(
				"ElapsedTime",
				elapsed.String(),
			),
		)
	})
}
