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

	"github.com/sourcenetwork/defradb/logging"
)

var (
	DAGSyncTimeout = time.Second * 60
)

// A DAGSyncer is an abstraction to an IPLD-based P2P storage layer.  A
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
	session     *sync.WaitGroup // A waitgroup to wait for all related jobs to conclude
	bp          *blockProcessor // the block processor to use
	cid         cid.Cid         // the cid of the block to fetch from the P2P network
	isComposite bool            // whether this is a composite block

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
			jobs, ok := docWorkerQueue[newJob.bp.dsKey.DocID]
			if !ok {
				jobs = make(chan *dagJob, numWorkers)
				for i := 0; i < numWorkers; i++ {
					go p.dagWorker(jobs)
				}
				docWorkerQueue[newJob.bp.dsKey.DocID] = jobs
			}
			jobs <- newJob

		case docID := <-p.closeJob:
			if jobs, ok := docWorkerQueue[docID]; ok {
				close(jobs)
				delete(docWorkerQueue, docID)
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
			logging.NewKV("Datastore Key", job.bp.dsKey),
			logging.NewKV("CID", job.cid),
		)

		select {
		case <-p.ctx.Done():
			// drain jobs from queue when we are done
			job.session.Done()
			continue
		default:
		}

		go func(j *dagJob) {
			if j.bp.getter != nil && j.cid.Defined() {
				cNode, err := j.bp.getter.Get(p.ctx, j.cid)
				if err != nil {
					log.ErrorE(p.ctx, "Failed to get node", err, logging.NewKV("CID", j.cid))
					j.session.Done()
					return
				}
				err = j.bp.processRemoteBlock(
					p.ctx,
					j.session,
					cNode,
					j.isComposite,
				)
				if err != nil {
					log.ErrorE(p.ctx, "Failed to process remote block", err, logging.NewKV("CID", j.cid))
				}
			}
			p.queuedChildren.Remove(j.cid)
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
