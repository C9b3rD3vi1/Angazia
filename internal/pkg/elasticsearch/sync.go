package elasticsearch

import (
	"context"
	"log"
	"time"
)

type SyncService struct {
	esClient         *ESClient
	jobIndexer       *JobIndexer
	candidateIndexer *CandidateIndexer
	companyIndexer   *CompanyIndexer
	jobCh            chan SyncEvent
	done             chan struct{}
	batchSize        int
	flushInterval    time.Duration
}

type SyncEvent struct {
	Action  string      // "index" or "delete"
	Index   string      // "jobs", "candidates", "companies"
	DocID   string
	Payload interface{} // *JobDocument, *CandidateDocument, *CompanyDocument for index; nil for delete
}

func NewSyncService(esClient *ESClient, jobIndexer *JobIndexer, candidateIndexer *CandidateIndexer, companyIndexer *CompanyIndexer) *SyncService {
	return &SyncService{
		esClient:         esClient,
		jobIndexer:       jobIndexer,
		candidateIndexer: candidateIndexer,
		companyIndexer:   companyIndexer,
		jobCh:            make(chan SyncEvent, 1000),
		done:             make(chan struct{}),
		batchSize:        50,
		flushInterval:    5 * time.Second,
	}
}

func (s *SyncService) Start(ctx context.Context) {
	go s.processQueue(ctx)
	log.Println("Elasticsearch sync service started")
}

func (s *SyncService) Stop() {
	close(s.done)
	log.Println("Elasticsearch sync service stopped")
}

func (s *SyncService) Enqueue(event SyncEvent) {
	select {
	case s.jobCh <- event:
	default:
		log.Printf("Sync queue full, dropping event: %s/%s/%s", event.Action, event.Index, event.DocID)
	}
}

func (s *SyncService) processQueue(ctx context.Context) {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	var batch []SyncEvent

	for {
		select {
		case <-s.done:
			s.flush(ctx, batch)
			return
		case <-ticker.C:
			if len(batch) > 0 {
				s.flush(ctx, batch)
				batch = batch[:0]
			}
		case event := <-s.jobCh:
			batch = append(batch, event)
			if len(batch) >= s.batchSize {
				s.flush(ctx, batch)
				batch = batch[:0]
			}
		}
	}
}

func (s *SyncService) flush(ctx context.Context, batch []SyncEvent) {
	if len(batch) == 0 {
		return
	}

	for _, event := range batch {
		switch event.Index {
		case "jobs":
			s.processJobEvent(ctx, event)
		case "candidates":
			s.processCandidateEvent(ctx, event)
		case "companies":
			s.processCompanyEvent(ctx, event)
		default:
			log.Printf("Unknown sync index: %s", event.Index)
		}
	}
}

func (s *SyncService) processJobEvent(ctx context.Context, event SyncEvent) {
	switch event.Action {
	case "index":
		if doc, ok := event.Payload.(*JobDocument); ok && doc != nil {
			if err := s.esClient.IndexJob(ctx, doc); err != nil {
				log.Printf("Failed to index job %s: %v", event.DocID, err)
			}
		}
	case "delete":
		if err := s.jobIndexer.DeleteJob(ctx, event.DocID); err != nil {
			log.Printf("Failed to delete job %s from ES: %v", event.DocID, err)
		}
	}
}

func (s *SyncService) processCandidateEvent(ctx context.Context, event SyncEvent) {
	switch event.Action {
	case "index":
		if doc, ok := event.Payload.(*CandidateDocument); ok && doc != nil {
			if err := s.esClient.IndexCandidate(ctx, doc); err != nil {
				log.Printf("Failed to index candidate %s: %v", event.DocID, err)
			}
		}
	case "delete":
		if err := s.candidateIndexer.DeleteCandidate(ctx, event.DocID); err != nil {
			log.Printf("Failed to delete candidate %s from ES: %v", event.DocID, err)
		}
	}
}

func (s *SyncService) processCompanyEvent(ctx context.Context, event SyncEvent) {
	switch event.Action {
	case "index":
		if doc, ok := event.Payload.(*CompanyDocument); ok && doc != nil {
			if err := s.esClient.IndexCompany(ctx, doc); err != nil {
				log.Printf("Failed to index company %s: %v", event.DocID, err)
			}
		}
	case "delete":
		if err := s.companyIndexer.DeleteCompany(ctx, event.DocID); err != nil {
			log.Printf("Failed to delete company %s from ES: %v", event.DocID, err)
		}
	}
}
