// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	DAGSyncTimeout = time.Second * 60
)

// A DAGSyncer is an abstraction to an IPLD-based p2p storage layer.  A
// DAGSyncer is a DAGService with the ability to publish new ipld nodes to the
// network, and retrieving others from it.
type DAGSyncer interface {
	ipld.DAGService
	// Returns true if the block is locally available (therefore, it
	// is considered processed).
	HasBlock(ctx context.Context, c cid.Cid) (bool, error)
}

// A SessionDAGSyncer is a Sessions-enabled DAGSyncer. This type of DAG-Syncer
// provides an optimized NodeGetter to make multiple related requests. The
// same session-enabled NodeGetter is used to download DAG branches when
// the DAGSyncer supports it.
type SessionDAGSyncer interface {
	DAGSyncer
	Session(context.Context) ipld.NodeGetter
}

type dagJob struct {
	session    *sync.WaitGroup // A waitgroup to wait for all related jobs to conclude
	nodeGetter ipld.NodeGetter // a node getter to use
	node       ipld.Node       // the current ipld Node

	collection client.Collection // collection our document belongs to
	dsKey      core.DataStoreKey // datastore key of our document
	fieldName  string            // field of the subgraph our node belongs to

	// Transaction common to a pushlog event. It is used to pass it along to processLog
	// and handleChildBlocks within the dagWorker.
	txn datastore.Txn

	// OLD FIELDS
	// root       cid.Cid         // the root of the branch we are walking down
	// rootPrio   uint64          // the priority of the root delta
	// delta      core.Delta      // the current delta
}

// the only purpose of this worker is to be able to orderly shut-down job
// workers without races by becoming the only sender for the store.jobQueue
// channel.
func (p *Peer) sendJobWorker() {
	// The DAG sync process for a document is handled over a single transaction, it is possible that a single
	// document ends up using all workers. Since the transaction uses a mutex to guarantee thread safety, some
	// operations in those workers may temporarily blocked which would leave a concurrent document sync process
	// hanging waiting for some workers to free up. To eliviate this problem, we add new workers dedicated to a
	// document and discard them once the process is completed.
	docWorkerQueue := make(map[string]chan *dagJob)
	for {
		select {
		case <-p.ctx.Done():
			for _, job := range docWorkerQueue {
				close(job)
			}
			return

		case newJob := <-p.sendJobs:
			jobs, ok := docWorkerQueue[newJob.dsKey.DocKey]
			if !ok {
				jobs = make(chan *dagJob, numWorkers)
				for i := 0; i < numWorkers; i++ {
					go p.dagWorker(jobs)
				}
				docWorkerQueue[newJob.dsKey.DocKey] = jobs
			}
			jobs <- newJob

		case dockey := <-p.closeJob:
			if jobs, ok := docWorkerQueue[dockey]; ok {
				close(jobs)
				delete(docWorkerQueue, dockey)
			}
		}
	}
}

// dagWorker should run in its own goroutine. Workers are launched during
// initialization in New().
func (p *Peer) dagWorker(jobs chan *dagJob) {
	for job := range jobs {
		log.Debug(
			p.ctx,
			"Starting new job from DAG queue",
			logging.NewKV("Datastore Key", job.dsKey),
			logging.NewKV("CID", job.node.Cid()),
		)

		select {
		case <-p.ctx.Done():
			// drain jobs from queue when we are done
			job.session.Done()
			continue
		default:
		}

		children, err := p.processLog(
			p.ctx,
			job.txn,
			job.collection,
			job.dsKey,
			job.node.Cid(),
			job.fieldName,
			job.node,
			job.nodeGetter,
			true,
		)
		if err != nil {
			log.ErrorE(
				p.ctx,
				"Error processing log",
				err,
				logging.NewKV("Datastore key", job.dsKey),
				logging.NewKV("CID", job.node.Cid()),
			)
			job.session.Done()
			continue
		}

		if len(children) == 0 {
			job.session.Done()
			continue
		}

		go func(j *dagJob) {
			p.handleChildBlocks(
				j.session,
				j.txn,
				j.collection,
				j.dsKey,
				j.fieldName,
				j.node,
				children,
				j.nodeGetter,
			)
			j.session.Done()
		}(job)
	}
}

type cidSafeSet struct {
	set map[cid.Cid]struct{}
	mux sync.Mutex
}

func newCidSafeSet() *cidSafeSet {
	return &cidSafeSet{
		set: make(map[cid.Cid]struct{}),
	}
}

// Visit checks if we can visit this node, or
// if its already being visited
func (s *cidSafeSet) Visit(c cid.Cid) bool {
	var b bool
	s.mux.Lock()
	{
		if _, ok := s.set[c]; !ok {
			s.set[c] = struct{}{}
			b = true
		}
	}
	s.mux.Unlock()
	return b
}

func (s *cidSafeSet) Remove(c cid.Cid) {
	s.mux.Lock()
	{
		delete(s.set, c)
	}
	s.mux.Unlock()
}
