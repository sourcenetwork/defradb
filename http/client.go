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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	sse "github.com/vito/go-sse/sse"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/http/graphql"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
)

var _ client.TxnStore = (*Client)(nil)

// SubscriptionTransport specifies the transport protocol to use for GraphQL subscriptions.
type SubscriptionTransport string

const (
	// SSESubscriptionTransport uses Server-Sent Events for subscriptions (default).
	SSESubscriptionTransport SubscriptionTransport = "sse"
	// WebSocketSubscriptionTransport uses WebSocket for subscriptions.
	WebSocketSubscriptionTransport SubscriptionTransport = "websocket"
)

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithSubscriptionTransport sets the transport protocol to use for GraphQL subscriptions.
func WithSubscriptionTransport(transport SubscriptionTransport) ClientOption {
	return func(c *Client) {
		c.subscriptionTransport = transport
	}
}

// Client implements the client.TxnStore interface over HTTP.
type Client struct {
	http                  *httpClient
	subscriptionTransport SubscriptionTransport
}

func NewClient(rawURL string, opts ...ClientOption) (*Client, error) {
	httpClient, err := newHttpClient(rawURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		http:                  httpClient,
		subscriptionTransport: SSESubscriptionTransport, // default to SSE
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) NewTxn(readOnly bool) (client.Txn, error) {
	query := url.Values{}
	if readOnly {
		query.Add("read_only", "true")
	}

	methodURL := c.http.apiURL.JoinPath("tx")
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var txRes CreateTxResponse
	if err := c.http.requestJson(req, &txRes); err != nil {
		return nil, err
	}
	return &Transaction{&Client{http: c.http, subscriptionTransport: c.subscriptionTransport}, txRes.ID}, nil
}

func (c *Client) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	query := url.Values{}
	if readOnly {
		query.Add("read_only", "true")
	}

	methodURL := c.http.apiURL.JoinPath("tx", "concurrent")
	methodURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var txRes CreateTxResponse
	if err := c.http.requestJson(req, &txRes); err != nil {
		return nil, err
	}
	return &Transaction{&Client{http: c.http, subscriptionTransport: c.subscriptionTransport}, txRes.ID}, nil
}

func (c *Client) BasicImport(ctx context.Context, filepath string) error {
	methodURL := c.http.apiURL.JoinPath("backup", "import")

	body, err := json.Marshal(&client.BackupConfig{Filepath: filepath})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	methodURL := c.http.apiURL.JoinPath("backup", "export")

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	methodURL := c.http.apiURL.JoinPath("schema")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), strings.NewReader(schema))
	if err != nil {
		return nil, err
	}
	var cols []client.CollectionVersion
	if err := c.http.requestJson(req, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

type patchCollectionRequest struct {
	Patch     string
	Migration immutable.Option[model.Lens]
}

func (c *Client) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
) error {
	methodURL := c.http.apiURL.JoinPath("collections")

	body, err := json.Marshal(patchCollectionRequest{patch, migration})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) SetActiveCollectionVersion(ctx context.Context, collectionVersionID string) error {
	methodURL := c.http.apiURL.JoinPath("collections", "default")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(),
		strings.NewReader(collectionVersionID))
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

type addViewRequest struct {
	Query        string
	SDL          string
	TransformCID immutable.Option[string]
}

func (c *Client) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	methodURL := c.http.apiURL.JoinPath("view")

	body, err := json.Marshal(addViewRequest{query, sdl, transformCID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	var descriptions []client.CollectionVersion
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}

	return descriptions, nil
}

func (c *Client) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	methodURL := c.http.apiURL.JoinPath("view", "refresh")
	params := url.Values{}
	if options.Name.HasValue() {
		params.Add("name", options.Name.Value())
	}
	if options.VersionID.HasValue() {
		params.Add("version_id", options.VersionID.Value())
	}
	if options.CollectionID.HasValue() {
		params.Add("collection_id", options.CollectionID.Value())
	}
	if options.IncludeInactive.HasValue() {
		params.Add("get_inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}
	methodURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}

func (c *Client) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	methodURL := c.http.apiURL.JoinPath("collections", "migrations")

	body, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	var res SetMigrationResponse
	if err := c.http.requestJson(req, &res); err != nil {
		return "", err
	}

	return res.LensID, nil
}

func (c *Client) AddLens(ctx context.Context, lens model.Lens) (string, error) {
	methodURL := c.http.apiURL.JoinPath("lens")

	body, err := json.Marshal(AddLensRequest{Lens: lens})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	var res AddLensResponse
	if err := c.http.requestJson(req, &res); err != nil {
		return "", err
	}

	return res.LensID, nil
}

func (c *Client) ListLenses(ctx context.Context) (map[string]model.Lens, error) {
	methodURL := c.http.apiURL.JoinPath("lens")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var res ListLensesResponse
	if err := c.http.requestJson(req, &res); err != nil {
		return nil, err
	}

	return res.Lenses, nil
}

func (c *Client) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	cols, err := c.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, NewErrCollectionNotFound(name)
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

func (c *Client) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	methodURL := c.http.apiURL.JoinPath("collections")
	params := url.Values{}
	if options.Name.HasValue() {
		params.Add("name", options.Name.Value())
	}
	if options.VersionID.HasValue() {
		params.Add("version_id", options.VersionID.Value())
	}
	if options.CollectionID.HasValue() {
		params.Add("collection_id", options.CollectionID.Value())
	}
	if options.IncludeInactive.HasValue() {
		params.Add("get_inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}
	methodURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var descriptions []client.CollectionVersion
	if err := c.http.requestJson(req, &descriptions); err != nil {
		return nil, err
	}
	collections := make([]client.Collection, len(descriptions))
	for i, d := range descriptions {
		collections[i] = &Collection{c.http, d}
	}
	return collections, nil
}

func (c *Client) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	methodURL := c.http.apiURL.JoinPath("collections", "indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes map[client.CollectionName][]client.IndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Client) ListAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	methodURL := c.http.apiURL.JoinPath("encrypted-indexes")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return nil, err
	}
	var indexes map[client.CollectionName][]client.EncryptedIndexDescription
	if err := c.http.requestJson(req, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (c *Client) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	methodURL := c.http.apiURL.JoinPath("graphql")
	result := &client.RequestResult{}

	gqlOptions := &client.GQLOptions{}
	for _, o := range opts {
		o(gqlOptions)
	}
	gqlRequest := &graphql.Request{
		Query:         query,
		OperationName: gqlOptions.OperationName,
		Variables:     gqlOptions.Variables,
	}

	body, err := json.Marshal(gqlRequest)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), bytes.NewBuffer(body))
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}

	err = c.http.setDefaultHeaders(req)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}

	op, err := parseGraphQLOperation(query)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}

	if op == ast.OperationTypeSubscription {
		// Use WebSocket transport if configured
		if c.subscriptionTransport == WebSocketSubscriptionTransport {
			return c.execRequestWebsocket(ctx, query, opts...)
		}
		// Default to SSE transport
		req.Header.Set("Accept", sseAcceptHeader)
	}

	res, err := c.http.client.Do(req)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	if res.Header.Get("Content-Type") == sseAcceptHeader {
		result.Subscription = c.execRequestSSE(res.Body)
		return result
	}
	// ignore close errors because they have
	// no perceivable effect on the end user
	// and cannot be reconciled easily
	defer res.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	if err = json.Unmarshal(data, &result.GQL); err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	return result
}

func (c *Client) execRequestSSE(r io.ReadCloser) chan client.GQLResult {
	resCh := make(chan client.GQLResult)
	go func() {
		eventReader := sse.NewReadCloser(r)
		defer func() {
			// ignore close errors because the status
			// and body of the request are already
			// checked and it cannot be handled properly
			eventReader.Close() //nolint:errcheck
			close(resCh)
		}()

		for {
			evt, err := eventReader.Next()
			if err != nil {
				return
			}
			var res client.GQLResult
			if err := json.Unmarshal(evt.Data, &res); err != nil {
				res.Errors = append(res.Errors, err)
			}
			resCh <- res
		}
	}()

	return resCh
}

// execRequestWebsocket executes a GraphQL subscription request over a WebSocket connection.
// This method only supports subscription operations. For queries and mutations, use ExecRequest.
func (c *Client) execRequestWebsocket(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	result := &client.RequestResult{}

	gqlOptions := &client.GQLOptions{}
	for _, o := range opts {
		o(gqlOptions)
	}

	wsURL := *c.http.apiURL
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	wsURL.Path = strings.TrimSuffix(wsURL.Path, "/") + "/graphql"

	dialer := websocket.Dialer{
		Subprotocols: []string{"graphql-ws"},
	}

	conn, _, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}

	initMsg := map[string]any{
		"type": "connection_init",
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		conn.Close()
		return result
	}

	// Wait for connection_ack
	var ackMsg map[string]any
	if err := conn.ReadJSON(&ackMsg); err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		conn.Close()
		return result
	}
	if ackMsg["type"] != "connection_ack" {
		result.GQL.Errors = append(result.GQL.Errors, ErrInvalidGraphQLRequest)
		conn.Close()
		return result
	}

	// Read and discard keep-alive message if present
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var kaMsg map[string]any
	if err := conn.ReadJSON(&kaMsg); err == nil {
		// ignore keep-alive messages
	}
	_ = conn.SetReadDeadline(time.Time{}) // Clear deadline

	gqlRequest := &graphql.Request{
		Query:         query,
		OperationName: gqlOptions.OperationName,
		Variables:     gqlOptions.Variables,
	}
	startMsg := map[string]any{
		"id":      "1",
		"type":    "start",
		"payload": gqlRequest,
	}
	if err := conn.WriteJSON(startMsg); err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		conn.Close()
		return result
	}

	result.Subscription = c.execWebsocketSubscription(ctx, conn)
	return result
}

func (c *Client) execWebsocketSubscription(ctx context.Context, conn *websocket.Conn) chan client.GQLResult {
	resCh := make(chan client.GQLResult)
	go func() {
		defer func() {
			// Send stop message before closing
			stopMsg := map[string]any{
				"id":   "1",
				"type": "stop",
			}
			conn.WriteJSON(stopMsg)
			conn.Close()
			close(resCh)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg map[string]any
				if err := conn.ReadJSON(&msg); err != nil {
					return
				}

				switch msg["type"] {
				case "data":
					var res client.GQLResult
					if payload, ok := msg["payload"].(map[string]any); ok {
						payloadBytes, _ := json.Marshal(payload)
						if err := json.Unmarshal(payloadBytes, &res); err != nil {
							res.Errors = append(res.Errors, err)
						}
					}
					resCh <- res
				case "error":
					var res client.GQLResult
					if payload, ok := msg["payload"].([]any); ok {
						for _, e := range payload {
							if errMap, ok := e.(map[string]any); ok {
								if errMsg, ok := errMap["message"].(string); ok {
									res.Errors = append(res.Errors, errors.New(errMsg))
								}
							}
						}
					}
					resCh <- res
				case "complete":
					return
				case "ka":
					// Keep-alive, ignore
				}
			}
		}
	}()

	return resCh
}

func (c *Client) PrintDump(ctx context.Context) error {
	methodURL := c.http.apiURL.JoinPath("debug", "dump")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) Purge(ctx context.Context) error {
	methodURL := c.http.apiURL.JoinPath("purge")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) HealthCheck(ctx context.Context) error {
	methodURL := c.http.baseURL.JoinPath("health-check")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = c.http.request(req)
	return err
}

func (c *Client) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	methodURL := c.http.apiURL.JoinPath("node", "identity")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	var ident immutable.Option[identity.PublicRawIdentity]
	if err := c.http.requestJson(req, &ident); err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return ident, err
}

func (c *Client) VerifySignature(ctx context.Context, cid string, pubKey crypto.PublicKey) error {
	methodURL := c.http.apiURL.JoinPath("block", "verify-signature")

	params := url.Values{}
	params.Add("cid", cid)
	params.Add("public-key", pubKey.String())
	params.Add("type", string(pubKey.Type()))
	methodURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, methodURL.String(), nil)
	if err != nil {
		return err
	}

	err = c.http.setDefaultHeaders(req)
	if err != nil {
		return err
	}

	_, err = c.http.request(req)
	return err
}

func parseGraphQLQuery(query string) (*ast.Document, error) {
	return parser.Parse(parser.ParseParams{
		Source: &source.Source{
			Body: []byte(query),
			Name: "GraphQL",
		},
	})
}

func parseGraphQLOperation(query string) (string, error) {
	doc, err := parseGraphQLQuery(query)
	if err != nil {
		return "", err
	}

	if len(doc.Definitions) == 0 {
		return "", ErrInvalidGraphQLRequest
	}

	op, ok := doc.Definitions[0].(*ast.OperationDefinition)
	if !ok {
		return "", ErrInvalidGraphQLRequest
	}

	return op.Operation, nil
}
