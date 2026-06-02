package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
)

type ESClient struct {
	client *elasticsearch.Client
	cfg    *config.Config
}

type JobDocument struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Requirements    string    `json:"requirements"`
	Responsibilities string   `json:"responsibilities"`
	RequiredSkills  []string  `json:"required_skills"`
	NiceToHaveSkills []string `json:"nice_to_have_skills"`
	ExperienceLevel string    `json:"experience_level"`
	MinExperience   int       `json:"min_experience"`
	MaxExperience   int       `json:"max_experience"`
	Location        string    `json:"location"`
	IsRemote        bool      `json:"is_remote"`
	IsHybrid        bool      `json:"is_hybrid"`
	EmploymentType  string    `json:"employment_type"`
	SalaryMin       int       `json:"salary_min"`
	SalaryMax       int       `json:"salary_max"`
	SalaryCurrency  string    `json:"salary_currency"`
	CompanyID       string    `json:"company_id"`
	CompanyName     string    `json:"company_name"`
	CompanyIndustry string    `json:"company_industry"`
	IsActive        bool      `json:"is_active"`
	IsFeatured      bool      `json:"is_featured"`
	ViewsCount      int       `json:"views_count"`
	ApplicationsCount int     `json:"applications_count"`
	PostedAt        time.Time `json:"posted_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CandidateDocument struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	FullName          string    `json:"full_name"`
	Headline          string    `json:"headline"`
	Bio               string    `json:"bio"`
	Skills            []string  `json:"skills"`
	ExperienceLevel   string    `json:"experience_level"`
	YearsOfExperience int       `json:"years_of_experience"`
	Location          string    `json:"location"`
	IsRemote          bool      `json:"is_remote"`
	IsHybrid          bool      `json:"is_hybrid"`
	GithubUsername    string    `json:"github_username"`
	GitHubActivity    GitHubActivityDoc `json:"github_activity"`
	IsAvailable       bool      `json:"is_available"`
	CreatedAt         time.Time `json:"created_at"`
}

type GitHubActivityDoc struct {
	PublicRepos   int      `json:"public_repos"`
	TotalCommits  int      `json:"total_commits"`
	Followers     int      `json:"followers"`
	TopLanguages  []string `json:"top_languages"`
	ActivityScore int      `json:"activity_score"`
}

type CompanyDocument struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Industry    string    `json:"industry"`
	Size        string    `json:"size"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`
	IsVerified  bool      `json:"is_verified"`
	Rating      float64   `json:"rating"`
	TotalJobs   int       `json:"total_jobs"`
	CreatedAt   time.Time `json:"created_at"`
}

type BulkDoc struct {
	ID   string
	Body interface{}
}

type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Score  float64         `json:"_score"`
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

func NewESClient(cfg *config.Config) (*ESClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Environment == "development"},
		},
	}
	
	if cfg.ElasticsearchUsername != "" && cfg.ElasticsearchPassword != "" {
		esCfg.Username = cfg.ElasticsearchUsername
		esCfg.Password = cfg.ElasticsearchPassword
	}
	
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}
	
	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %s", res.String())
	}
	
	log.Println("✅ Elasticsearch connected successfully")
	
	return &ESClient{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *ESClient) IndexJob(ctx context.Context, doc *JobDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	
	req := esapi.IndexRequest{
		Index:      "jobs",
		DocumentID: doc.ID,
		Body:       bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("ES index error: %s", res.String())
	}
	
	return nil
}

func (c *ESClient) IndexCandidate(ctx context.Context, doc *CandidateDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	
	req := esapi.IndexRequest{
		Index:      "candidates",
		DocumentID: doc.ID,
		Body:       bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("ES index error: %s", res.String())
	}
	
	return nil
}

func (c *ESClient) IndexCompany(ctx context.Context, doc *CompanyDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	
	req := esapi.IndexRequest{
		Index:      "companies",
		DocumentID: doc.ID,
		Body:       bytes.NewReader(body),
	}
	
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("ES index error: %s", res.String())
	}
	
	return nil
}

func (c *ESClient) DeleteDocument(ctx context.Context, index, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
	}
	
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("ES delete error: %s", res.String())
	}
	
	return nil
}

func (c *ESClient) Search(ctx context.Context, index string, query map[string]interface{}) (*SearchResponse, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}
	
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return nil, fmt.Errorf("ES search error: %s", res.String())
	}
	
	var result SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

func (c *ESClient) BulkIndex(ctx context.Context, index string, docs []BulkDoc) error {
	var buf bytes.Buffer
	
	for _, doc := range docs {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    doc.ID,
			},
		}
		
		metaJSON, _ := json.Marshal(meta)
		docJSON, _ := json.Marshal(doc.Body)
		
		buf.Write(metaJSON)
		buf.WriteByte('\n')
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}
	
	res, err := c.client.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("ES bulk error: %s", res.String())
	}
	
	return nil
}

func (c *ESClient) CreateIndex(ctx context.Context, index string, mapping string) error {
	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader([]byte(mapping)),
	}
	
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() && res.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("ES create index error: %s", res.String())
	}
	
	return nil
}