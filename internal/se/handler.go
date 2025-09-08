package se

/*type seHandler struct {
	peer *Peer
}

func (p *seHandler) handleSEQuery(req QuerySEArtifactsRequest) {
	p.server.mu.Lock()
	reps, exists := p.server.replicators[req.CollectionID]
	p.server.mu.Unlock()

	if !exists || len(reps) == 0 {
		req.Response <- se.QuerySEArtifactsResponse{
			DocIDs: []string{},
			Error:  nil,
		}
		return
	}

	grpcQueries := make([]seFieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		grpcQueries[i] = seFieldQuery{
			FieldName: q.FieldName,
			IndexID:   q.IndexID,
			SearchTag: q.SearchTag,
		}
	}

	grpcReq := querySEArtifactsRequest{
		CollectionID: req.CollectionID,
		Queries:      grpcQueries,
	}

	docIDSet := make(map[string]struct{})
	var queryErr error

	// TODO: ask replicators one-by-one.
	for pid := range reps {
		reply, err := p.server.querySEArtifacts(p.ctx, pid, grpcReq)
		if err != nil {
			log.ErrorE(
				"Failed querying SE artifacts from replicator",
				err,
				corelog.String("CollectionID", req.CollectionID),
				corelog.Any("PeerID", pid))
			queryErr = err
			continue
		}

		for _, docID := range reply.DocIDs {
			docIDSet[docID] = struct{}{}
		}
	}

	docIDs := make([]string, 0, len(docIDSet))
	for docID := range docIDSet {
		docIDs = append(docIDs, docID)
	}

	req.Response <- se.QuerySEArtifactsResponse{
		DocIDs: docIDs,
		Error:  queryErr,
	}
}

func (p *seHandler) pushSEArtifactsToReplicators(evt se.ReplicateEvent) {
	p.server.mu.Lock()
	reps, exists := p.server.replicators[evt.CollectionID]
	p.server.mu.Unlock()

	if exists {
		for pid := range reps {
			go func(peerID peer.ID) {
				if err := p.server.pushSEArtifacts(evt, peerID); err != nil {
					log.ErrorE(
						"Failed pushing SE artifacts",
						err,
						corelog.String("DocID", evt.DocID),
						corelog.String("CollectionID", evt.CollectionID),
						corelog.Any("PeerID", peerID))
				}
			}(pid)
		}
	}
}*/
