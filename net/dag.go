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
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

var (
	DAGSyncTimeout = time.Second * 60
)

type dagJob struct {
	session *sync.WaitGroup // A waitgroup to wait for all related jobs to conclude
	bp      *blockProcessor // the block processor to use
	cid     cid.Cid         // the cid of the block to fetch from the P2P network

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
		select {
		case <-p.ctx.Done():
			// drain jobs from queue when we are done
			job.session.Done()
			continue
		default:
		}

		go func(j *dagJob) {
			if j.bp.dagSyncer != nil && j.cid.Defined() {
				// BlockOfType will return the block if it is already in the store or fetch it from the network
				// if it is not. This is a blocking call and will wait for the block to be fetched.
				// It uses the LinkSystem to fetch the block. Blocks retrieved from the network will
				// also be stored in the blockstore in the same call.
				// Blocks have to match the coreblock.SchemaPrototype to be returned.
				nd, err := j.bp.dagSyncer.BlockOfType(p.ctx, cidlink.Link{Cid: j.cid}, coreblock.SchemaPrototype)
				if err != nil {
					log.ErrorContextE(
						p.ctx,
						"Failed to get node",
						err,
						corelog.Any("CID", j.cid))
					j.session.Done()
					return
				}
				block, err := coreblock.GetFromNode(nd)
				if err != nil {
					log.ErrorContextE(
						p.ctx,
						"Failed to convert ipld node to block",
						err,
						corelog.Any("CID", j.cid))
				}
				j.bp.handleChildBlocks(
					p.ctx,
					j.session,
					block,
				)
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
