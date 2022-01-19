package net

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document/key"
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

	txn        core.MultiStore   // txn our job is running within
	collection client.Collection // collection our document belongs to
	dockey     key.DocKey        // dockey of our document
	fieldName  string            // field of the subgraph our node belongs to

	// OLD FIELDS
	// root       cid.Cid         // the root of the branch we are walking down
	// rootPrio   uint64          // the priority of the root delta
	// delta      core.Delta      // the current delta
}

// handleBlock takes care of vetting, retrieving and applying
// CRDT blocks to the Datastore.
// func (peer *Peer) handleBlock(c cid.Cid) error {
// 	// Ignore already known blocks.
// 	// This includes the case when the block is a current
// 	// head.
// 	known, err := peer.ds.HasBlock(peer.ctx, c)
// 	if err != nil {
// 		return fmt.Errorf("error checking for known block: %w", err)
// 	}
// 	if known {
// 		log.Debugf("%s is known. Skip walking tree", c)
// 		return nil
// 	}

// 	// Walk down from this block.
// 	ctx, cancel := context.WithCancel(peer.ctx)
// 	defer cancel()

// 	dg := &clock.CrdtNodeGetter{NodeGetter: peer.ds}
// 	if sessionMaker, ok := peer.ds.(SessionDAGSyncer); ok {
// 		dg = &clock.CrdtNodeGetter{NodeGetter: sessionMaker.Session(ctx)}
// 	}

// 	var session sync.WaitGroup
// 	peer.sendNewJobs(&session, dg, c, 0, []cid.Cid{c})
// 	session.Wait()
// 	return nil
// }

// // sendNewJobs calls getDeltas (GetMany) on the crdtNodeGetter with the given
// // children and sends each response to the workers. It will block until all
// // jobs have been queued.
// func (p *Peer) sendNewJobs(session *sync.WaitGroup, ng *clock.CrdtNodeGetter, root cid.Cid, rootPrio uint64, children []cid.Cid) {
// 	if len(children) == 0 {
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(p.ctx, DAGSyncTimeout)
// 	defer cancel()

// 	// Special case for root
// 	if rootPrio == 0 {
// 		prio, err := ng.GetPriority(ctx, children[0])
// 		if err != nil {
// 			log.Errorf("error getting root delta priority: %s", err)
// 			return
// 		}
// 		rootPrio = prio
// 	}

// 	// This gets deltas but it is unable to tells us which childrens
// 	// failed to be fetched though.
// 	for deltaOpt := range ng.GetDeltas(ctx, children) {
// 		if deltaOpt.Error() != nil {
// 			log.Errorf("error getting delta: %s", deltaOpt.Error())
// 			continue
// 		}

// 		session.Add(1)
// 		job := &dagJob{
// 			session:    session,
// 			nodeGetter: ng,
// 			root:       root,
// 			delta:      deltaOpt.GetDelta(),
// 			node:       deltaOpt.GetNode(),
// 			rootPrio:   rootPrio,
// 		}
// 		select {
// 		case p.sendJobs <- job:
// 		case <-p.ctx.Done():
// 			// the job was never sent, so it cannot complete.
// 			session.Done()
// 			return
// 		}
// 	}

// 	// Clear up any children we failed to get from queued children
// 	for _, child := range children {
// 		p.queuedChildren.Remove(child)
// 	}
// }

// the only purpose of this worker is to be able to orderly shut-down job
// workers without races by becoming the only sender for the store.jobQueue
// channel.
func (p *Peer) sendJobWorker() {
	for {
		select {
		case <-p.ctx.Done():
			close(p.jobQueue)
			return
		case j := <-p.sendJobs:
			p.jobQueue <- j
		}
	}
}

// dagWorker should run in its own goroutine. Workers are launched during
// initialization in New().
func (p *Peer) dagWorker() {
	for job := range p.jobQueue {
		log.Debugf("Starting new job from dag queue: %s at %s", job.dockey, job.node.Cid())
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
			job.dockey,
			job.node.Cid(),
			job.fieldName,
			job.node,
			job.nodeGetter,
		)

		if err != nil {
			log.Errorf("Error processing log for %s at %s: %s", job.dockey, job.node.Cid(), err)
			job.session.Done()
			continue
		}
		go func(j *dagJob) {
			p.handleChildBlocks(j.txn, j.session, j.collection, j.dockey, job.fieldName, job.node, children, j.nodeGetter)
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
