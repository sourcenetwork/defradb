// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

type schemaVersionID = string

// LensDoc represents a document that will be sent to/from a Lens.
type LensDoc = map[string]any

type lensInput struct {
	SchemaVersionID schemaVersionID
	Doc             LensDoc
}

// Lens migrate items fed in to the target schema version.
//
// Source documents may be of various schema versions, and may need to be migrated across multiple
// versions.  As the input versions are unknown until enumerated, the migration pipeline is constructed
// lazily, as new source schema versions are discovered.  If a migration does not exist for a schema
// version, the document will be passed on to the next stage.
type Lens interface {
	enumerable.Enumerable[LensDoc]

	// Put feeds the given document into the Lens, so that its transformed output may be yielded.
	Put(schemaVersionID schemaVersionID, value LensDoc) error
}

type lens struct {
	lensRegistry client.LensRegistry

	// The primary access points to the lens pipes through which all things flow.
	lensPipesBySchemaVersionIDs map[schemaVersionID]enumerable.Concatenation[LensDoc]

	// The input pipes, into which items are added to the pipe system.
	lensInputPipesBySchemaVersionIDs map[schemaVersionID]enumerable.Queue[LensDoc]

	// The output pipe, through which all outputs must exit.
	outputPipe         enumerable.Concatenation[LensDoc]
	unknownVersionPipe enumerable.Queue[LensDoc]

	schemaVersionHistory map[schemaVersionID]*targetedSchemaHistoryLink

	source enumerable.Queue[lensInput]
}

var _ Lens = (*lens)(nil)

func new(
	lensRegistry client.LensRegistry,
	targetSchemaVersionID schemaVersionID,
	schemaVersionHistory map[schemaVersionID]*targetedSchemaHistoryLink,
) Lens {
	targetSource := enumerable.NewQueue[LensDoc]()
	outputPipe := enumerable.Concat[LensDoc](targetSource)

	return &lens{
		lensRegistry:         lensRegistry,
		source:               enumerable.NewQueue[lensInput](),
		outputPipe:           outputPipe,
		unknownVersionPipe:   targetSource,
		schemaVersionHistory: schemaVersionHistory,
		lensInputPipesBySchemaVersionIDs: map[schemaVersionID]enumerable.Queue[LensDoc]{
			targetSchemaVersionID: targetSource,
		},
		lensPipesBySchemaVersionIDs: map[schemaVersionID]enumerable.Concatenation[LensDoc]{
			targetSchemaVersionID: outputPipe,
		},
	}
}

// todo - instead of this and a lens-fetcher, we could instead make lens-fetcher (and other fetchers) enumerables
// instead and use those as the `source` directly.
// https://github.com/sourcenetwork/defradb/issues/1589
func (l *lens) Put(schemaVersionID schemaVersionID, value LensDoc) error {
	return l.source.Put(lensInput{
		SchemaVersionID: schemaVersionID,
		Doc:             value,
	})
}

// Next reads documents from source, and migrates them to the target schema version.
//
// Source documents may be of various schema versions, and may need to be migrated across multiple
// versions.  As the input versions are unknown until enumerated, the migration pipeline is constructed
// lazily, as new source schema versions are discovered.  If a migration does not exist for a schema
// version, the document will be passed on to the next stage.
//
// Perhaps the best way to visualize this is as a multi-input marble-run, where inputs and their paths
// are constructed as new marble types are discovered.
//
//   - Each version can have one or none migrations.
//   - Each migration in the document's path to the target version is guaranteed to recieve the document
//     exactly once.
//   - Schema history is assumed to be a single straight line with no branching, this will be fixed with
//     https://github.com/sourcenetwork/defradb/issues/1598
func (l *lens) Next() (bool, error) {
	// Check the output pipe first, there could be items remaining within waiting to be yielded.
	hasValue, err := l.outputPipe.Next()
	if err != nil || hasValue {
		return hasValue, err
	}

	hasValue, err = l.source.Next()
	if err != nil || !hasValue {
		return false, err
	}

	doc, err := l.source.Value()
	if err != nil {
		return false, err
	}

	var inputPipe enumerable.Queue[LensDoc]
	if p, ok := l.lensInputPipesBySchemaVersionIDs[doc.SchemaVersionID]; ok {
		// If the input pipe exists we can safely assume that it has been correctly connected
		// up to the output via any intermediary pipes.
		inputPipe = p
	} else {
		historyLocation, ok := l.schemaVersionHistory[doc.SchemaVersionID]
		if !ok {
			// We may recieve documents of unknown schema versions, they should
			// still be fed through the pipe system in order to preserve order.
			err = l.unknownVersionPipe.Put(doc.Doc)
			if err != nil {
				return false, err
			}

			return l.outputPipe.Next()
		}

		var pipeHead enumerable.Enumerable[LensDoc]

		for {
			junctionPipe, junctionPreviouslyExisted := l.lensPipesBySchemaVersionIDs[historyLocation.schemaVersionID]
			if !junctionPreviouslyExisted {
				versionInputPipe := enumerable.NewQueue[LensDoc]()
				l.lensInputPipesBySchemaVersionIDs[historyLocation.schemaVersionID] = versionInputPipe
				if inputPipe == nil {
					// The input pipe will be fed documents which are currently at this schema version
					inputPipe = versionInputPipe
				}
				// It is a source of the schemaVersion junction pipe, other schema versions
				// may also join as sources to this junction pipe
				junctionPipe = enumerable.Concat[LensDoc](versionInputPipe)
				l.lensPipesBySchemaVersionIDs[historyLocation.schemaVersionID] = junctionPipe
			}

			// If we have previously laid pipe, we need to connect it to the current junction.
			// This links a lens migration to the next stage.
			if pipeHead != nil {
				junctionPipe.Append(pipeHead)
			}

			if junctionPreviouslyExisted {
				// If the junction pipe previously existed, then we can assume it is already connected to outputPipe
				// via any intermediary pipes.
				break
			}

			if historyLocation.targetVector > 0 {
				// Aquire a lens migration from the registery, using the junctionPipe as its source.
				// The new pipeHead will then be connected as a source to the next migration-stage on
				// the next loop.
				pipeHead, err = l.lensRegistry.MigrateUp(junctionPipe, historyLocation.schemaVersionID)
				if err != nil {
					return false, err
				}

				historyLocation = historyLocation.next.Value()
			} else {
				// The pipe head then becomes the schema version migration to the next version
				// sourcing from any documents at schemaVersionID, or lower schema versions.
				// This also ensures each document only passes through each migration once,
				// in order, and through the same state container (in case migrations use state).
				pipeHead, err = l.lensRegistry.MigrateDown(junctionPipe, historyLocation.schemaVersionID)
				if err != nil {
					return false, err
				}

				// Aquire a lens migration from the registery, using the junctionPipe as its source.
				// The new pipeHead will then be connected as a source to the next migration-stage on
				// the next loop.
				historyLocation = historyLocation.previous.Value()
			}
		}
	}

	// Place the current doc in the appropriate input pipe
	err = inputPipe.Put(doc.Doc)
	if err != nil {
		return false, err
	}

	// Then draw out the next result result from the output pipe, pulling it through any migrations
	// along the way.  Typically this will be the (now migrated) document just placed into the input pipe.
	return l.outputPipe.Next()
}

func (l *lens) Value() (LensDoc, error) {
	return l.outputPipe.Value()
}

func (l *lens) Reset() {
	l.outputPipe.Reset()
}
