// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"

	blockstore "github.com/ipfs/boxo/blockstore"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

type CollectionName = string

// DB is the primary public programmatic access point to the local DefraDB instance.
//
// It should be constructed via the [db] package, via the [db.NewDB] function.
type DB interface {
	// Store contains DefraDB functions protected by an internal, short-lived, transaction, allowing safe
	// access to common database read and write operations.
	Store

	// NewTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is not threadsafe.
	NewTxn(context.Context, bool) (datastore.Txn, error)

	// NewConcurrentTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is threadsafe and multiple threads/Go routines
	// can safely operate on it concurrently.
	NewConcurrentTxn(context.Context, bool) (datastore.Txn, error)

	// WithTxn returns a new [client.Store] that respects the given transaction.
	WithTxn(datastore.Txn) Store

	// Root returns the underlying root store, within which all data managed by DefraDB is held.
	Root() datastore.RootStore

	// Blockstore returns the blockstore, within which all blocks (commits) managed by DefraDB are held.
	//
	// It sits within the rootstore returned by [Root].
	Blockstore() blockstore.Blockstore

	// Peerstore returns the peerstore where known host information is stored.
	//
	// It sits within the rootstore returned by [Root].
	Peerstore() datastore.DSBatching

	// Close closes the database instance and releases any resources held.
	//
	// The behaviour of other functions in this package after this function has been called is undefined
	// unless explicitly stated on the function in question.
	//
	// It does not explicitly clear any data from persisted storage, and a new [DB] instance may typically
	// be created after calling this to resume operations on the prior data - this is however dependant on
	// the behaviour of the rootstore provided on database instance creation, as this function will Close
	// the provided rootstore.
	Close()

	// Events returns the database event queue.
	//
	// It may be used to monitor database events - a new event will be yielded for each mutation.
	// Note: it does not copy the queue, just the reference to it.
	Events() events.Events

	// MaxTxnRetries returns the number of retries that this DefraDB instance has been configured to
	// make in the event of a transaction conflict in certain scenarios.
	//
	// Currently this is only used within the P2P system and will not affect operations initiated by users.
	MaxTxnRetries() int

	// PrintDump logs the entire contents of the rootstore (all the data managed by this DefraDB instance).
	//
	// It is likely unwise to call this on a large database instance.
	PrintDump(ctx context.Context) error
}

// Store contains the core DefraDB read-write operations.
type Store interface {
	// Backup holds the backup related methods that must be implemented by the database.
	Backup

	// AddSchema takes the provided GQL schema in SDL format, and applies it to the [Store],
	// creating the necessary collections, request types, etc.
	//
	// All schema types provided must not exist prior to calling this, and they may not reference existing
	// types previously defined.
	AddSchema(context.Context, string) ([]CollectionDescription, error)

	// PatchSchema takes the given JSON patch string and applies it to the set of SchemaDescriptions
	// present in the database.
	//
	// If true is provided, the new schema versions will be made active and previous versions deactivated, otherwise
	// [SetActiveSchemaVersion] should be called to do so.
	//
	// It will also update the GQL types used by the query system. It will error and not apply any of the
	// requested, valid updates should the net result of the patch result in an invalid state.  The
	// individual operations defined in the patch do not need to result in a valid state, only the net result
	// of the full patch.
	//
	// The collections (including the schema version ID) will only be updated if any changes have actually
	// been made, if the net result of the patch matches the current persisted description then no changes
	// will be applied.
	//
	// Field [FieldKind] values may be provided in either their raw integer form, or as string as per
	// [FieldKindStringToEnumMapping].
	//
	// A lens configuration may also be provided, it will be added to all collections using the schema.
	PatchSchema(context.Context, string, immutable.Option[model.Lens], bool) error

	// SetActiveSchemaVersion activates all collection versions with the given schema version, and deactivates all
	// those without it (if they share the same schema root).
	//
	// This will affect all operations interacting with the schema where a schema version is not explicitly
	// provided.  This includes GQL queries and Collection operations.
	//
	// It will return an error if the provided schema version ID does not exist.
	SetActiveSchemaVersion(context.Context, string) error

	// AddView creates a new Defra View.
	//
	// It takes a GQL query string, for example:
	//
	// Author {
	//	 name
	//	 books {
	//	   name
	//	 }
	// }
	//
	//
	// A GQL SDL that matches its output type must also be provided.  There can only be one `type` declaration,
	// any nested objects must be declared as embedded/schema-only types using the `interface` keyword.
	// Relations must only be specified on the parent side of the relationship.  For example:
	//
	// type AuthorView {
	//   name: String
	//   books: [BookView]
	// }
	// interface BookView {
	//   name: String
	// }
	//
	// It will return the collection definitions of the types defined in the SDL if successful, otherwise an error
	// will be returned.  This function does not execute the given query.
	AddView(ctx context.Context, gqlQuery string, sdl string) ([]CollectionDefinition, error)

	// SetMigration sets the migration for all collections using the given source-destination schema version IDs.
	//
	// There may only be one migration per collection version.  If another migration was registered it will be
	// overwritten by this migration.
	//
	// Neither of the schema version IDs specified in the configuration need to exist at the time of calling.
	// This is to allow the migration of documents of schema versions unknown to the local node received by the
	// P2P system.
	//
	// Migrations will only run if there is a complete path from the document schema version to the latest local
	// schema version.
	SetMigration(context.Context, LensConfig) error

	// LensRegistry returns the LensRegistry in use by this database instance.
	//
	// It exposes several useful thread-safe migration related functions.
	LensRegistry() LensRegistry

	// GetCollectionByName attempts to retrieve a collection matching the given name.
	//
	// If no matching collection is found an error will be returned.
	GetCollectionByName(context.Context, CollectionName) (Collection, error)

	// GetCollectionsBySchemaRoot attempts to retrieve all collections using the given schema ID.
	//
	// If no matching collection is found an empty set will be returned.
	GetCollectionsBySchemaRoot(context.Context, string) ([]Collection, error)

	// GetCollectionsByVersionID attempts to retrieve all collections using the given schema version ID.
	//
	// If no matching collections are found an empty set will be returned.
	GetCollectionsByVersionID(context.Context, string) ([]Collection, error)

	// GetAllCollections returns all collections and their descriptions that currently exist within
	// this [Store].
	//
	// If `true` is provided, the results will include inactive collections.  If `false`, only active collections
	// will be returned.
	GetAllCollections(context.Context, bool) ([]Collection, error)

	// GetSchemasByName returns the all schema versions with the given name.
	GetSchemasByName(context.Context, string) ([]SchemaDescription, error)

	// GetSchemaByVersionID returns the schema description for the schema version of the
	// ID provided.
	//
	// Will return an error if it is not found.
	GetSchemaByVersionID(context.Context, string) (SchemaDescription, error)

	// GetSchemasByRoot returns the all schema versions for the given root.
	GetSchemasByRoot(context.Context, string) ([]SchemaDescription, error)

	// GetAllSchemas returns all schema versions that currently exist within
	// this [Store].
	GetAllSchemas(context.Context) ([]SchemaDescription, error)

	// GetAllIndexes returns all the indexes that currently exist within this [Store].
	GetAllIndexes(context.Context) (map[CollectionName][]IndexDescription, error)

	// ExecRequest executes the given GQL request against the [Store].
	ExecRequest(context.Context, string) *RequestResult
}

// GQLResult represents the immediate results of a GQL request.
//
// It does not handle subscription channels. This object and its children are json serializable.
type GQLResult struct {
	// Errors contains any errors generated whilst attempting to execute the request.
	//
	// If there are values in this slice the request will likely not have run to completion
	// and [Data] will be nil.
	Errors []error `json:"errors,omitempty"`

	// Data contains the resultant data produced by the GQL request.
	//
	// It will be nil if any errors were raised during execution.
	Data any `json:"data"`
}

// RequestResult represents the results of a GQL request.
type RequestResult struct {
	// GQL contains the immediate results of the GQL request.
	GQL GQLResult

	// Pub contains a pointer to an event stream which channels any subscription results
	// if the request was a GQL subscription.
	Pub *events.Publisher[events.Update]
}
