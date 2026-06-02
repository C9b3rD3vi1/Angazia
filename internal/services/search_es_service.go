package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/elasticsearch"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

var (
	ErrSearchServiceNotAvailable = errors.New("search service not available")
	ErrESClientNotInitialized    = errors.New("elasticsearch client not initialized")
)

type SearchESService struct {
	esClient *elasticsearch.ESClient
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

func NewSearchESService(esClient *elasticsearch.ESClient, jobRepo repository.JobRepository, userRepo repository.UserRepository) *SearchESService {
	return &SearchESService{
		esClient: esClient,
		jobRepo:  jobRepo,
		userRepo: userRepo,
	}
}

func (s *SearchESService) IsAvailable() bool {
	return s.esClient != nil
}

type JobSearchResult struct {
	Jobs    []*models.Job
	Total   int64
	Page    int
	Limit   int
	Aggs    map[string]interface{}
}

type CandidateSearchResult struct {
	Candidates []*models.EmployeeProfile
	Total      int64
	Page       int
	Limit      int
	Aggs       map[string]interface{}
}

type CompanySearchResult struct {
	Companies []*models.EmployerProfile
	Total     int64
	Page      int
	Limit     int
	Aggs      map[string]interface{}
}

func (s *SearchESService) SearchJobs(ctx context.Context, filters map[string]interface{}, page, limit int) (*JobSearchResult, error) {
	if s.esClient == nil {
		return nil, ErrESClientNotInitialized
	}

	indexer := elasticsearch.NewJobIndexer(s.esClient, s.jobRepo, s.userRepo)
	query := indexer.BuildSearchQuery(filters, page, limit, false)

	resp, err := s.esClient.Search(ctx, "jobs", query)
	if err != nil {
		return nil, err
	}

	result := &JobSearchResult{
		Total: int64(resp.Hits.Total.Value),
		Page:  page,
		Limit: limit,
		Aggs:  resp.Aggregations,
	}

	for _, hit := range resp.Hits.Hits {
		var doc elasticsearch.JobDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}

		job, err := s.jobRepo.GetByID(ctx, doc.ID)
		if err != nil {
			continue
		}
		result.Jobs = append(result.Jobs, job)
	}

	return result, nil
}

func (s *SearchESService) SearchCandidates(ctx context.Context, filters elasticsearch.CandidateSearchFilters) (*CandidateSearchResult, error) {
	if s.esClient == nil {
		return nil, ErrESClientNotInitialized
	}

	resp, err := s.esClient.SearchCandidates(ctx, filters)
	if err != nil {
		return nil, err
	}

	result := &CandidateSearchResult{
		Total: int64(len(resp.Hits)),
		Page:  filters.Page,
		Limit: filters.Limit,
		Aggs:  resp.Aggs,
	}

	for _, hit := range resp.Hits {
		profile, err := s.userRepo.GetEmployeeProfile(ctx, hit.Source.UserID)
		if err != nil {
			continue
		}
		result.Candidates = append(result.Candidates, profile)
	}

	return result, nil
}

func (s *SearchESService) SearchCompanies(ctx context.Context, filters elasticsearch.CompanySearchFilters) (*CompanySearchResult, error) {
	if s.esClient == nil {
		return nil, ErrESClientNotInitialized
	}

	resp, err := s.esClient.SearchCompanies(ctx, filters)
	if err != nil {
		return nil, err
	}

	result := &CompanySearchResult{
		Total: int64(len(resp.Hits)),
		Page:  filters.Page,
		Limit: filters.Limit,
		Aggs:  resp.Aggs,
	}

	for _, hit := range resp.Hits {
		profile, err := s.userRepo.GetEmployerProfile(ctx, hit.Source.ID)
		if err != nil {
			continue
		}
		result.Companies = append(result.Companies, profile)
	}

	return result, nil
}

func (s *SearchESService) BuildJobSearchQuery(filters map[string]interface{}, page, limit int) map[string]interface{} {
	return elasticsearch.NewJobIndexer(s.esClient, s.jobRepo, s.userRepo).BuildSearchQuery(filters, page, limit, false)
}
