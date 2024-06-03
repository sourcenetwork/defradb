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

	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/sourcenetwork/corelog"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

var (
	dagSyncTimeout = time.Second * 60
)

type blockProcessor struct {
	*Peer
	wg             *sync.WaitGroup
	bsSession      *blockservice.Session
	queuedChildren *sync.Map
}

func newBlockProcessor(
	ctx context.Context,
	p *Peer,
) *blockProcessor {
	return &blockProcessor{
		Peer:           p,
		wg:             &sync.WaitGroup{},
		bsSession:      blockservice.NewSession(ctx, p.bserv),
		queuedChildren: &sync.Map{},
	}
}

// processRemoteBlock stores the block in the DAG store and initiates a sync of the block's children.
func (bp *blockProcessor) processRemoteBlock(
	ctx context.Context,
	block *coreblock.Block,
) error {
	// Store the block in the DAG store
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(bp.db.Blockstore().AsIPLDStorage())
	_, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}
	// Initiate a sync of the block's children
	bp.wg.Add(1)
	bp.handleChildBlocks(ctx, block)

	return nil
}

func (bp *blockProcessor) handleChildBlocks(
	ctx context.Context,
	block *coreblock.Block,
) {
	defer bp.wg.Done()

	if len(block.Links) == 0 {
		return
	}

	links := make([]cid.Cid, 0, len(block.Links))
	for _, link := range block.Links {
		exists, err := bp.db.Blockstore().Has(ctx, link.Cid)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to check if block exists",
				err,
				corelog.Any("CID", link.Cid),
			)
			continue
		}
		if exists {
			continue
		}
		if _, loaded := bp.queuedChildren.LoadOrStore(link.Cid, struct{}{}); !loaded {
			links = append(links, link.Cid)
		}
	}

	getCtx, cancel := context.WithTimeout(ctx, dagSyncTimeout)
	defer cancel()

	childBlocks := bp.bsSession.GetBlocks(getCtx, links)

	for rawBlock := range childBlocks {
		block, err := coreblock.GetFromBytes(rawBlock.RawData())
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to get block from bytes",
				err,
				corelog.Any("CID", rawBlock.Cid()),
			)
			continue
		}
		bp.wg.Add(1)
		go bp.handleChildBlocks(ctx, block)
	}

	for _, link := range links {
		bp.queuedChildren.Delete(link)
	}
}
