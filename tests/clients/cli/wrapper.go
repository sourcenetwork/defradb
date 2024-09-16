// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.DB = (*Wrapper)(nil)

type Wrapper struct {
	node       *node.Node
	cmd        *cliWrapper
	handler    *http.Handler
	httpServer *httptest.Server
}

// NewWrapper takes a Node, and a SourceHub address used to pay for SourceHub transactions.
//
// sourceHubAddress can (and will) be empty when testing non sourceHub ACP implementations.
func NewWrapper(node *node.Node, sourceHubAddress string) (*Wrapper, error) {
	handler, err := http.NewHandler(node.DB)
	if err != nil {
		return nil, err
	}

	httpServer := httptest.NewServer(handler)
	cmd := newCliWrapper(httpServer.URL, sourceHubAddress)

	return &Wrapper{
		node:       node,
		cmd:        cmd,
		httpServer: httpServer,
		handler:    handler,
	}, nil
}

func (w *Wrapper) PeerInfo() peer.AddrInfo {
	args := []string{"client", "p2p", "info"}

	data, err := w.cmd.execute(context.Background(), args)
	if err != nil {
		panic(fmt.Sprintf("failed to get peer info: %v", err))
	}
	var info peer.AddrInfo
	if err := json.Unmarshal(data, &info); err != nil {
		panic(fmt.Sprintf("failed to get peer info: %v", err))
	}
	return info
}

func (w *Wrapper) SetReplicator(ctx context.Context, rep client.Replicator) error {
	args := []string{"client", "p2p", "replicator", "set"}
	args = append(args, "--collection", strings.Join(rep.Schemas, ","))

	info, err := json.Marshal(rep.Info)
	if err != nil {
		return err
	}
	args = append(args, string(info))

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	args := []string{"client", "p2p", "replicator", "delete"}
	args = append(args, "--collection", strings.Join(rep.Schemas, ","))

	info, err := json.Marshal(rep.Info)
	if err != nil {
		return err
	}
	args = append(args, string(info))

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	args := []string{"client", "p2p", "replicator", "getall"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var reps []client.Replicator
	if err := json.Unmarshal(data, &reps); err != nil {
		return nil, err
	}
	return reps, nil
}

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	args := []string{"client", "p2p", "collection", "add"}
	args = append(args, strings.Join(collectionIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	args := []string{"client", "p2p", "collection", "remove"}
	args = append(args, strings.Join(collectionIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	args := []string{"client", "p2p", "collection", "getall"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var cols []string
	if err := json.Unmarshal(data, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	args := []string{"client", "backup", "import"}
	args = append(args, filepath)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	args := []string{"client", "backup", "export"}

	if len(config.Collections) > 0 {
		args = append(args, "--collections", strings.Join(config.Collections, ","))
	}
	if config.Format != "" {
		args = append(args, "--format", config.Format)
	}
	if config.Pretty {
		args = append(args, "--pretty")
	}
	args = append(args, config.Filepath)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) AddPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	args := []string{"client", "acp", "policy", "add"}
	args = append(args, policy)

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	var addPolicyResult client.AddPolicyResult
	if err := json.Unmarshal(data, &addPolicyResult); err != nil {
		return client.AddPolicyResult{}, err
	}

	return addPolicyResult, err
}

func (w *Wrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
	args := []string{"client", "schema", "add"}
	args = append(args, schema)

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var cols []client.CollectionDescription
	if err := json.Unmarshal(data, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (w *Wrapper) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setDefault bool,
) error {
	args := []string{"client", "schema", "patch"}
	if setDefault {
		args = append(args, "--set-active")
	}
	args = append(args, patch)

	if migration.HasValue() {
		lenses, err := json.Marshal(migration.Value())
		if err != nil {
			return err
		}
		args = append(args, string(lenses))
	}

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
) error {
	args := []string{"client", "collection", "patch"}
	args = append(args, patch)
	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	args := []string{"client", "schema", "set-active"}
	args = append(args, schemaVersionID)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	args := []string{"client", "view", "add"}
	args = append(args, query)
	args = append(args, sdl)

	if transform.HasValue() {
		lenses, err := json.Marshal(transform.Value())
		if err != nil {
			return nil, err
		}
		args = append(args, string(lenses))
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var defs []client.CollectionDefinition
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}
	return defs, nil
}

func (w *Wrapper) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	args := []string{"client", "view", "refresh"}
	if options.Name.HasValue() {
		args = append(args, "--name", options.Name.Value())
	}
	if options.SchemaVersionID.HasValue() {
		args = append(args, "--version", options.SchemaVersionID.Value())
	}
	if options.SchemaRoot.HasValue() {
		args = append(args, "--schema", options.SchemaRoot.Value())
	}
	if options.IncludeInactive.HasValue() {
		args = append(args, "--get-inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	args := []string{"client", "schema", "migration", "set"}

	lenses, err := json.Marshal(config.Lens)
	if err != nil {
		return err
	}
	args = append(args, config.SourceSchemaVersionID)
	args = append(args, config.DestinationSchemaVersionID)
	args = append(args, string(lenses))

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) LensRegistry() client.LensRegistry {
	return &LensRegistry{w.cmd}
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	cols, err := w.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	args := []string{"client", "collection", "describe"}
	if options.Name.HasValue() {
		args = append(args, "--name", options.Name.Value())
	}
	if options.SchemaVersionID.HasValue() {
		args = append(args, "--version", options.SchemaVersionID.Value())
	}
	if options.SchemaRoot.HasValue() {
		args = append(args, "--schema", options.SchemaRoot.Value())
	}
	if options.IncludeInactive.HasValue() {
		args = append(args, "--get-inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var colDesc []client.CollectionDefinition
	if err := json.Unmarshal(data, &colDesc); err != nil {
		return nil, err
	}
	cols := make([]client.Collection, len(colDesc))
	for i, v := range colDesc {
		cols[i] = &Collection{w.cmd, v}
	}
	return cols, err
}

func (w *Wrapper) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	schemas, err := w.GetSchemas(ctx, client.SchemaFetchOptions{ID: immutable.Some(versionID)})
	if err != nil {
		return client.SchemaDescription{}, err
	}

	// schemas will always have length == 1 here
	return schemas[0], nil
}

func (w *Wrapper) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	args := []string{"client", "schema", "describe"}
	if options.ID.HasValue() {
		args = append(args, "--version", options.ID.Value())
	}
	if options.Root.HasValue() {
		args = append(args, "--root", options.Root.Value())
	}
	if options.Name.HasValue() {
		args = append(args, "--name", options.Name.Value())
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var schema []client.SchemaDescription
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}
	return schema, err
}

func (w *Wrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes map[client.CollectionName][]client.IndexDescription
	if err := json.Unmarshal(data, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	args := []string{"client", "query"}
	args = append(args, query)

	options := &client.GQLOptions{}
	for _, o := range opts {
		o(options)
	}

	result := &client.RequestResult{}
	if options.OperationName != "" {
		args = append(args, "--operation", options.OperationName)
	}
	if len(options.Variables) > 0 {
		enc, err := json.Marshal(options.Variables)
		if err != nil {
			result.GQL.Errors = []error{err}
			return result
		}
		args = append(args, "--variables", string(enc))
	}

	stdOut, stdErr, err := w.cmd.executeStream(ctx, args)
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	buffer := bufio.NewReader(stdOut)
	header, err := buffer.ReadString('\n')
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	if header == cli.SUB_RESULTS_HEADER {
		result.Subscription = w.execRequestSubscription(buffer)
		return result
	}
	data, err := io.ReadAll(buffer)
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	errData, err := io.ReadAll(stdErr)
	if err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	if len(errData) > 0 {
		result.GQL.Errors = []error{fmt.Errorf("%s", errData)}
		return result
	}

	var response http.GraphQLResponse
	if err = json.Unmarshal(data, &response); err != nil {
		result.GQL.Errors = []error{err}
		return result
	}
	result.GQL.Data = response.Data
	result.GQL.Errors = response.Errors
	return result
}

func (w *Wrapper) execRequestSubscription(r io.Reader) chan client.GQLResult {
	resCh := make(chan client.GQLResult)
	go func() {
		dec := json.NewDecoder(r)
		defer close(resCh)

		for {
			var response http.GraphQLResponse
			if err := dec.Decode(&response); err != nil {
				return
			}
			resCh <- client.GQLResult{
				Errors: response.Errors,
				Data:   response.Data,
			}
		}
	}()
	return resCh
}

func (w *Wrapper) NewTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	args := []string{"client", "tx", "create"}
	if readOnly {
		args = append(args, "--read-only")
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var res http.CreateTxResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	tx, err := w.handler.Transaction(res.ID)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx, w.cmd}, nil
}

func (w *Wrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	args := []string{"client", "tx", "create"}
	args = append(args, "--concurrent")

	if readOnly {
		args = append(args, "--read-only")
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var res http.CreateTxResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	tx, err := w.handler.Transaction(res.ID)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx, w.cmd}, nil
}

func (w *Wrapper) Rootstore() datastore.Rootstore {
	return w.node.DB.Rootstore()
}

func (w *Wrapper) Blockstore() datastore.Blockstore {
	return w.node.DB.Blockstore()
}

func (w *Wrapper) Headstore() ds.Read {
	return w.node.DB.Headstore()
}

func (w *Wrapper) Peerstore() datastore.DSBatching {
	return w.node.DB.Peerstore()
}

func (w *Wrapper) Close() {
	w.httpServer.CloseClientConnections()
	w.httpServer.Close()
	_ = w.node.Close(context.Background())
}

func (w *Wrapper) Events() *event.Bus {
	return w.node.DB.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.node.DB.PrintDump(ctx)
}

func (w *Wrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	return w.node.Peer.Connect(ctx, addr)
}

func (w *Wrapper) Host() string {
	return w.httpServer.URL
}
