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
	"time"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

type Wrapper struct {
	node         *node.Node
	cmd          *cliWrapper
	handler      *http.Handler
	httpServer   *httptest.Server
	serverCancel context.CancelFunc
}

// NewWrapper takes a Node, and a SourceHub address used to pay for SourceHub transactions.
//
// sourceHubAddress can (and will) be empty when testing non sourceHub ACP implementations.
func NewWrapper(node *node.Node, sourceHubAddress string) (*Wrapper, error) {
	handler, err := http.NewHandler(node.DB)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	handlerWithCtx := http.InjectServerContext(ctx)(handler)
	httpServer := httptest.NewServer(handlerWithCtx)
	cmd := newCliWrapper(httpServer.URL, sourceHubAddress)

	return &Wrapper{
		node:         node,
		cmd:          cmd,
		httpServer:   httpServer,
		handler:      handler,
		serverCancel: cancel,
	}, nil
}

func (w *Wrapper) PeerInfo() ([]string, error) {
	args := []string{"client", "p2p", "info"}

	data, err := w.cmd.execute(context.Background(), args)
	if err != nil {
		return nil, err
	}
	var addresses []string
	if err := json.Unmarshal(data, &addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (w *Wrapper) ActivePeers(ctx context.Context) ([]string, error) {
	args := []string{"client", "p2p", "active-peers"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var peers []string
	if err := json.Unmarshal(data, &peers); err != nil {
		return nil, err
	}
	return peers, nil
}

func (w *Wrapper) Connect(ctx context.Context, addresses []string) error {
	args := []string{"client", "p2p", "connect"}

	args = append(args, strings.Join(addresses, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) SetReplicator(ctx context.Context, addresses []string, collections ...string) error {
	args := []string{"client", "p2p", "replicator", "set"}
	args = append(args, "--collection", strings.Join(collections, ","))

	args = append(args, strings.Join(addresses, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, id string, collections ...string) error {
	args := []string{"client", "p2p", "replicator", "delete"}
	args = append(args, "--collection", strings.Join(collections, ","))

	args = append(args, id)

	_, err := w.cmd.execute(ctx, args)
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

func (w *Wrapper) CreateP2PCollections(ctx context.Context, collectionIDs ...string) error {
	args := []string{"client", "p2p", "collection", "create"}
	args = append(args, strings.Join(collectionIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) DeleteP2PCollections(ctx context.Context, collectionIDs ...string) error {
	args := []string{"client", "p2p", "collection", "delete"}
	args = append(args, strings.Join(collectionIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) ListP2PCollections(ctx context.Context) ([]string, error) {
	args := []string{"client", "p2p", "collection", "list"}

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

func (w *Wrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	args := []string{"client", "p2p", "document", "add"}
	args = append(args, strings.Join(docIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	args := []string{"client", "p2p", "document", "remove"}
	args = append(args, strings.Join(docIDs, ","))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	args := []string{"client", "p2p", "document", "getall"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var docIDs []string
	if err := json.Unmarshal(data, &docIDs); err != nil {
		return nil, err
	}
	return docIDs, nil
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	args := []string{"client", "p2p", "document", "sync"}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		args = append(args, "--timeout", time.Until(deadline).String())
	}

	args = append(args, collectionName)
	args = append(args, docIDs...)

	_, err := w.cmd.execute(context.Background(), args)
	return err
}

func (w *Wrapper) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	args := []string{"client", "p2p", "collection", "sync-versions"}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		args = append(args, "--timeout", time.Until(deadline).String())
	}

	args = append(args, versionIDs...)

	_, err := w.cmd.execute(context.Background(), args)
	return err
}

func (w *Wrapper) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	args := []string{"client", "p2p", "collection", "sync-branchable", collectionID}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		args = append(args, "--timeout", time.Until(deadline).String())
	}

	_, err := w.cmd.execute(ctx, args)
	return err
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

func (w *Wrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	args := []string{"client", "schema", "add"}
	args = append(args, schema)

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var cols []client.CollectionVersion
	if err := json.Unmarshal(data, &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
) error {
	args := []string{"client", "collection", "patch"}
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

func (w *Wrapper) SetActiveCollectionVersion(ctx context.Context, collectionVersionID string) error {
	args := []string{"client", "collection", "set-active"}
	args = append(args, collectionVersionID)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	args := []string{"client", "view", "add"}
	args = append(args, "--query", query)
	args = append(args, "--sdl", sdl)

	if transformCID.HasValue() {
		args = append(args, "--lens-cid", transformCID.Value())
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var defs []client.CollectionVersion
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
	if options.VersionID.HasValue() {
		args = append(args, "--version-id", options.VersionID.Value())
	}
	if options.CollectionID.HasValue() {
		args = append(args, "--collection-id", options.CollectionID.Value())
	}
	if options.IncludeInactive.HasValue() {
		args = append(args, "--get-inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	args := []string{"client", "lens", "set"}

	lenses, err := json.Marshal(config.Lens)
	if err != nil {
		return "", err
	}
	args = append(args, config.SourceCollectionVersionID)
	args = append(args, config.DestinationCollectionVersionID)
	args = append(args, string(lenses))

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return "", err
	}

	var lensID string
	if err := json.Unmarshal(data, &lensID); err != nil {
		return "", err
	}
	return lensID, nil
}

func (w *Wrapper) AddLens(ctx context.Context, lens model.Lens) (string, error) {
	args := []string{"client", "lens", "add"}

	lensJSON, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	args = append(args, string(lensJSON))

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return "", err
	}

	var lensID string
	if err := json.Unmarshal(data, &lensID); err != nil {
		return "", err
	}
	return lensID, nil
}

func (w *Wrapper) ListLenses(ctx context.Context) (map[string]model.Lens, error) {
	args := []string{"client", "lens", "list"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}

	var lenses map[string]model.Lens
	if err := json.Unmarshal(data, &lenses); err != nil {
		return nil, err
	}
	return lenses, nil
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
	if options.VersionID.HasValue() {
		args = append(args, "--version-id", options.VersionID.Value())
	}
	if options.CollectionID.HasValue() {
		args = append(args, "--collection-id", options.CollectionID.Value())
	}
	if options.IncludeInactive.HasValue() {
		args = append(args, "--get-inactive", strconv.FormatBool(options.IncludeInactive.Value()))
	}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var colDesc []client.CollectionVersion
	if err := json.Unmarshal(data, &colDesc); err != nil {
		return nil, err
	}
	cols := make([]client.Collection, len(colDesc))
	for i, v := range colDesc {
		cols[i] = &Collection{w.cmd, v}
	}
	return cols, err
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

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	args := []string{"client", "encrypted-index", "list"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes map[client.CollectionName][]client.EncryptedIndexDescription
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
			result.GQL.Errors = append(result.GQL.Errors, err)
			return result
		}
		args = append(args, "--variables", string(enc))
	}

	stdOut, stdErr, err := w.cmd.executeStream(ctx, args)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	buffer := bufio.NewReader(stdOut)
	header, err := buffer.ReadString('\n')
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	if header == cli.SUB_RESULTS_HEADER {
		result.Subscription = w.execRequestSubscription(buffer)
		return result
	}
	data, err := io.ReadAll(buffer)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	errData, err := io.ReadAll(stdErr)
	if err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
		return result
	}
	if len(errData) > 0 {
		result.GQL.Errors = append(result.GQL.Errors, fmt.Errorf("%s", errData))
		return result
	}

	if err = json.Unmarshal(data, &result.GQL); err != nil {
		result.GQL.Errors = append(result.GQL.Errors, err)
	}
	return result
}

func (w *Wrapper) execRequestSubscription(r io.Reader) chan client.GQLResult {
	resCh := make(chan client.GQLResult)
	go func() {
		dec := json.NewDecoder(r)
		defer close(resCh)

		for {
			var res client.GQLResult
			if err := dec.Decode(&res); err != nil {
				res.Errors = append(res.Errors, err)
			}
			resCh <- res
		}
	}()
	return resCh
}

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	args := []string{"client", "tx", "create"}
	if readOnly {
		args = append(args, "--read-only")
	}

	data, err := w.cmd.execute(context.Background(), args)
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
	return &Transaction{w, tx}, nil
}

func (w *Wrapper) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	args := []string{"client", "tx", "create"}
	args = append(args, "--concurrent")

	if readOnly {
		args = append(args, "--read-only")
	}

	data, err := w.cmd.execute(context.Background(), args)
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
	return &Transaction{w, tx}, nil
}

func (w *Wrapper) Close() {
	w.serverCancel()
	w.httpServer.Close()
	_ = w.node.Close(context.Background())
}

func (w *Wrapper) Events() event.Bus {
	return w.node.DB.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	args := []string{"dump"}

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) Host() string {
	return w.httpServer.URL
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	args := []string{"client", "node-identity"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	var res identity.PublicRawIdentity
	if err := json.Unmarshal(data, &res); err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(res), nil
}

func (w *Wrapper) VerifySignature(ctx context.Context, cid string, pubKey crypto.PublicKey) error {
	args := []string{"client", "block", "verify-signature"}

	args = append(args, "--type", string(pubKey.Type()))
	args = append(args, pubKey.String(), cid)

	_, err := w.cmd.execute(ctx, args)
	return err
}
