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
	"strconv"
	"time"

	"github.com/sourcenetwork/defradb/logging"
)

type logger struct {
	logging.Logger
}

func defaultLogger() *logger {
	return &logger{
		Logger: logging.MustNewLogger("defra.http"),
	}
}

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

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.ResponseWriter.Header().Get("Content-Length") != "" {
		return lrw.ResponseWriter.Write(b)
	}
	lrw.contentLength = len(b)
	lrw.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(lrw.contentLength))
	return lrw.ResponseWriter.Write(b)
}

func (l *logger) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		elapsed := time.Since(start)
		l.Info(
			r.Context(),
			"Request",
			logging.NewKV(
				"Method",
				r.Method,
			),
			logging.NewKV(
				"Path",
				r.URL.Path,
			),
			logging.NewKV(
				"Status",
				lrw.statusCode,
			),
			logging.NewKV(
				"Length",
				lrw.contentLength,
			),
			logging.NewKV(
				"Elapsed",
				elapsed,
			),
		)
	})
}
