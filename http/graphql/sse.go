// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	// "github.com/vektah/gqlparser/gqlerror"
)

type (
	SSE struct {
		KeepAlivePingInterval time.Duration
	}

	sseConnection struct {
		ctx             context.Context
		mu              sync.Mutex
		f               http.Flusher
		keepAliveTicker *time.Ticker
	}
)

var _ Transport = SSE{}

func getRequestBody(r *http.Request) (string, error) {
	if r == nil || r.Body == nil {
		return "", nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("unable to get Request Body %w", err)
	}
	return string(body), nil
}

func (t SSE) Supports(r *http.Request) bool {
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return false
	}
	return r.Method == http.MethodPost && mediaType == "application/json"
}

func (t SSE) Methods() []string {
	return []string{http.MethodPost}
}

func (t SSE) Do(w http.ResponseWriter, r *http.Request, executer Executor) {
	ctx := r.Context()
	flusher, ok := w.(http.Flusher)
	if !ok {
		err := errors.New("streaming unsupported")
		responseJSON(w, http.StatusBadRequest, errorResponse{err})
		return
	}

	c := &sseConnection{
		ctx: ctx,
		f:   flusher,
	}

	defer c.flush()

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "application/json")

	bodyString, err := getRequestBody(r)
	if err != nil {
		gqlErr := errors.Wrap("could not get json request body: %w", err)
		responseJSON(w, http.StatusBadRequest, errorResponse{gqlErr})
		return
	}

	var request Request
	bodyReader := strings.NewReader(bodyString)
	err = json.NewDecoder(bodyReader).Decode(&request)
	if err != nil {
		responseJSON(w, http.StatusBadRequest, errorResponse{err})
		return
	}

	var options []client.RequestOption
	if request.OperationName != "" {
		options = append(options, client.WithOperationName(request.OperationName))
	}
	if len(request.Variables) > 0 {
		options = append(options, client.WithVariables(request.Variables))
	}

	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprint(w, ":\n\n")
	c.flush()

	if t.KeepAlivePingInterval > 0 {
		c.mu.Lock()
		c.keepAliveTicker = time.NewTicker(t.KeepAlivePingInterval)
		c.mu.Unlock()

		go c.keepAlive(w)
	}

	// responses, ctx := exec.DispatchOperation(ctx, rc)
	result := executer(r.Context(), request.Query, options...)
	if len(result.GQL.Errors) > 0 {
		writeResultWithSSE(w, result.GQL) // todo
	}

	for {
		select {
		// unified context for client and server shutdown
		case <-r.Context().Done():
			fmt.Fprint(w, "event: complete\n\n")
			return
		case item, open := <-result.Subscription:
			if !open {
				fmt.Fprint(w, "event: complete\n\n")
				return
			}

			writeResultWithSSE(w, item)
		}

		c.flush()
		c.resetTicker(t.KeepAlivePingInterval)
	}
}

func (c *sseConnection) resetTicker(interval time.Duration) {
	if interval != 0 {
		c.mu.Lock()
		c.keepAliveTicker.Reset(interval)
		c.mu.Unlock()
	}
}

func (c *sseConnection) keepAlive(w io.Writer) {
	for {
		select {
		case <-c.ctx.Done():
			c.keepAliveTicker.Stop()
			return
		case <-c.keepAliveTicker.C:
			fmt.Fprintf(w, ": ping\n\n")
			c.flush()
		}
	}
}

func (c *sseConnection) flush() {
	c.mu.Lock()
	c.f.Flush()
	c.mu.Unlock()
}

func writeJsonWithSSE(w io.Writer, response Response) {
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "event: next\ndata: %s\n\n", b)
}

func writeResultWithSSE(w io.Writer, result client.GQLResult) {
	b, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "event: next\ndata: %s\n\n", b)
}
