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
	"strings"

	blockstore "github.com/ipfs/boxo/blockstore"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/http"
)

var _ client.DB = (*Wrapper)(nil)

type Wrapper struct {
	db         client.DB
	store      client.Store
	cmd        *cliWrapper
	handler    *http.Handler
	httpServer *httptest.Server
}

func NewWrapper(db client.DB) *Wrapper {
	handler := http.NewHandler(db, http.ServerOptions{})
	httpServer := httptest.NewServer(handler)
	cmd := newCliWrapper(httpServer.URL)

	// TODO use http.Wrapper here to make a lot of this obsolete

	return &Wrapper{
		db:         db,
		store:      db,
		cmd:        cmd,
		httpServer: httpServer,
		handler:    handler,
	}
}

func (w *Wrapper) SetReplicator(ctx context.Context, rep client.Replicator) error {
	args := []string{"client", "p2p", "replicator", "set"}
	args = append(args, "--collection", strings.Join(rep.Schemas, ","))

	addrs, err := peer.AddrInfoToP2pAddrs(&rep.Info)
	if err != nil {
		return err
	}
	args = append(args, addrs[0].String())

	_, err = w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	args := []string{"client", "p2p", "replicator", "delete"}

	addrs, err := peer.AddrInfoToP2pAddrs(&rep.Info)
	if err != nil {
		return err
	}
	args = append(args, addrs[0].String())

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

func (w *Wrapper) AddP2PCollection(ctx context.Context, collectionID string) error {
	args := []string{"client", "p2p", "collection", "add"}
	args = append(args, collectionID)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	args := []string{"client", "p2p", "collection", "remove"}
	args = append(args, collectionID)

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

func (w *Wrapper) PatchSchema(ctx context.Context, patch string) error {
	args := []string{"client", "schema", "patch"}
	args = append(args, patch)

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	return w.LensRegistry().SetMigration(ctx, config)
}

func (w *Wrapper) LensRegistry() client.LensRegistry {
	return &lensWrapper{
		lens: w.store.LensRegistry(),
		cmd:  w.cmd,
	}
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	return w.store.GetCollectionByName(ctx, name)
}

func (w *Wrapper) GetCollectionBySchemaID(ctx context.Context, schemaId string) (client.Collection, error) {
	return w.store.GetCollectionBySchemaID(ctx, schemaId)
}

func (w *Wrapper) GetCollectionByVersionID(ctx context.Context, versionId string) (client.Collection, error) {
	return w.store.GetCollectionByVersionID(ctx, versionId)
}

func (w *Wrapper) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	return w.store.GetAllCollections(ctx)
}

func (w *Wrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}

	data, err := w.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var index map[client.CollectionName][]client.IndexDescription
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return index, nil
}

func (w *Wrapper) ExecRequest(ctx context.Context, query string) *client.RequestResult {
	args := []string{"client", "query"}
	args = append(args, query)

	result := &client.RequestResult{}

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
	if header == SUB_RESULTS_HEADER {
		result.Pub = w.execRequestSubscription(ctx, buffer)
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
	for _, err := range response.Errors {
		result.GQL.Errors = append(result.GQL.Errors, fmt.Errorf(err))
	}
	return result
}

func (w *Wrapper) execRequestSubscription(ctx context.Context, r io.Reader) *events.Publisher[events.Update] {
	pubCh := events.New[events.Update](0, 0)
	pub, err := events.NewPublisher[events.Update](pubCh, 0)
	if err != nil {
		return nil
	}

	go func() {
		dec := json.NewDecoder(r)

		for {
			var response http.GraphQLResponse
			if err := dec.Decode(&response); err != nil {
				return
			}
			var errors []error
			for _, err := range response.Errors {
				errors = append(errors, fmt.Errorf(err))
			}
			pub.Publish(client.GQLResult{
				Errors: errors,
				Data:   response.Data,
			})
		}
	}()

	return pub
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
	return &TxWrapper{tx, w.cmd}, nil
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
	return &TxWrapper{tx, w.cmd}, nil
}

func (w *Wrapper) WithTxn(tx datastore.Txn) client.Store {
	return &Wrapper{
		db:    w.db,
		store: w.db.WithTxn(tx),
		cmd:   w.cmd.withTxn(tx),
	}
}

func (w *Wrapper) Root() datastore.RootStore {
	return w.db.Root()
}

func (w *Wrapper) Blockstore() blockstore.Blockstore {
	return w.db.Blockstore()
}

func (w *Wrapper) Close(ctx context.Context) {
	w.httpServer.CloseClientConnections()
	w.httpServer.Close()
	w.db.Close(ctx)
}

func (w *Wrapper) Events() events.Events {
	return w.db.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.db.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.db.PrintDump(ctx)
}
