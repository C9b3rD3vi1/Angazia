package elasticsearch

import (
	"context"
	"log"
	"sort"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type JobIndexer struct {
	esClient *ESClient
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

type CandidateIndexer struct {
	esClient *ESClient
	userRepo repository.UserRepository
}

type CompanyIndexer struct {
	esClient *ESClient
	userRepo repository.UserRepository
}

func NewJobIndexer(esClient *ESClient, jobRepo repository.JobRepository, userRepo repository.UserRepository) *JobIndexer {
	return &JobIndexer{
		esClient: esClient,
		jobRepo:  jobRepo,
		userRepo: userRepo,
	}
}

func NewCandidateIndexer(esClient *ESClient, userRepo repository.UserRepository) *CandidateIndexer {
	return &CandidateIndexer{
		esClient: esClient,
		userRepo: userRepo,
	}
}

func NewCompanyIndexer(esClient *ESClient, userRepo repository.UserRepository) *CompanyIndexer {
	return &CompanyIndexer{
		esClient: esClient,
		userRepo: userRepo,
	}
}

func (i *JobIndexer) IndexJob(ctx context.Context, job *models.Job) error {
	employer, err := i.userRepo.GetEmployerProfile(ctx, job.EmployerID)
	if err != nil {
		return err
	}

	doc := &JobDocument{
		ID:                 job.ID,
		Title:              job.Title,
		Description:        job.Description,
		Requirements:       job.Requirements,
		Responsibilities:   job.Responsibilities,
		RequiredSkills:     job.RequiredSkills,
		NiceToHaveSkills:   job.NiceToHaveSkills,
		ExperienceLevel:    job.ExperienceLevel,
		MinExperience:      job.MinExperience,
		MaxExperience:      job.MaxExperience,
		Location:           job.Location,
		IsRemote:           job.IsRemote,
		IsHybrid:           job.IsHybrid,
		EmploymentType:     job.EmploymentType,
		SalaryMin:          job.SalaryMin,
		SalaryMax:          job.SalaryMax,
		SalaryCurrency:     job.SalaryCurrency,
		CompanyID:          employer.UserID,
		CompanyName:        employer.CompanyName,
		CompanyIndustry:    employer.Industry,
		IsActive:           job.IsActive,
		IsFeatured:         job.IsFeatured,
		ViewsCount:         job.ViewsCount,
		ApplicationsCount:  job.ApplicationsCount,
		PostedAt:           job.PostedAt,
		UpdatedAt:          job.PostedAt,
	}

	return i.esClient.IndexJob(ctx, doc)
}

func (i *JobIndexer) IndexAllJobs(ctx context.Context) error {
	log.Println("Indexing all jobs to Elasticsearch...")

	page := 1
	limit := 100
	totalIndexed := 0

	for {
		jobs, total, err := i.jobRepo.List(ctx, repository.JobFilters{}, page, limit)
		if err != nil {
			return err
		}

		employerIDs := make([]string, 0, len(jobs))
		employerMap := make(map[string]*models.EmployerProfile)

		for _, job := range jobs {
			employerIDs = append(employerIDs, job.EmployerID)
		}

		if err := i.userRepo.GetEmployerProfilesBatch(ctx, employerIDs, employerMap); err != nil {
			log.Printf("Failed to load employers in batch: %v, falling back to individual queries", err)
		}

		docs := make([]BulkDoc, 0, len(jobs))
		for _, job := range jobs {
			employer, ok := employerMap[job.EmployerID]
			if !ok {
				var eErr error
				employer, eErr = i.userRepo.GetEmployerProfile(ctx, job.EmployerID)
				if eErr != nil {
					log.Printf("Failed to load employer for job %s: %v", job.ID, eErr)
					continue
				}
			}

			doc := &JobDocument{
				ID:                 job.ID,
				Title:              job.Title,
				Description:        job.Description,
				Requirements:       job.Requirements,
				Responsibilities:   job.Responsibilities,
				RequiredSkills:     job.RequiredSkills,
				NiceToHaveSkills:   job.NiceToHaveSkills,
				ExperienceLevel:    job.ExperienceLevel,
				MinExperience:      job.MinExperience,
				MaxExperience:      job.MaxExperience,
				Location:           job.Location,
				IsRemote:           job.IsRemote,
				IsHybrid:           job.IsHybrid,
				EmploymentType:     job.EmploymentType,
				SalaryMin:          job.SalaryMin,
				SalaryMax:          job.SalaryMax,
				SalaryCurrency:     job.SalaryCurrency,
				CompanyID:          employer.UserID,
				CompanyName:        employer.CompanyName,
				CompanyIndustry:    employer.Industry,
				IsActive:           job.IsActive,
				IsFeatured:         job.IsFeatured,
				ViewsCount:         job.ViewsCount,
				ApplicationsCount:  job.ApplicationsCount,
				PostedAt:           job.PostedAt,
				UpdatedAt:          job.PostedAt,
			}
			docs = append(docs, BulkDoc{ID: job.ID, Body: doc})
		}

		if len(docs) > 0 {
			if err := i.esClient.BulkIndex(ctx, "jobs", docs); err != nil {
				log.Printf("Failed to bulk index %d jobs: %v", len(docs), err)
				for _, d := range docs {
					if doc, ok := d.Body.(*JobDocument); ok {
						if e := i.esClient.IndexJob(ctx, doc); e != nil {
							log.Printf("Failed to index job %s: %v", d.ID, e)
						} else {
							totalIndexed++
						}
					}
				}
			} else {
				totalIndexed += len(docs)
			}
		}

		if int64(page*limit) >= total {
			break
		}
		page++
	}

	log.Printf("Indexed %d jobs to Elasticsearch", totalIndexed)
	return nil
}

func (i *JobIndexer) DeleteJob(ctx context.Context, jobID string) error {
	return i.esClient.DeleteDocument(ctx, "jobs", jobID)
}

func (i *JobIndexer) BuildSearchQuery(filters map[string]interface{}, page, limit int, includeInactive bool) map[string]interface{} {
	from := (page - 1) * limit

	query := map[string]interface{}{
		"from": from,
		"size": limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []interface{}{},
				"filter": []interface{}{},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_score": map[string]string{"order": "desc"},
			},
			map[string]interface{}{
				"posted_at": map[string]string{"order": "desc"},
			},
		},
		"aggs": BuildJobAggregations(),
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     keyword,
				"fields":    []string{"title^3", "description", "requirements", "required_skills^2"},
				"fuzziness": "AUTO",
			},
		})
	}

	if location, ok := filters["location"].(string); ok && location != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"location": location,
			},
		})
	}

	if isRemote, ok := filters["is_remote"].(bool); ok {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"is_remote": isRemote,
			},
		})
	}

	if skills, ok := filters["skills"].([]string); ok && len(skills) > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"terms": map[string]interface{}{
				"required_skills": skills,
			},
		})
	}

	if minSalary, ok := filters["min_salary"].(float64); ok && minSalary > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"salary_max": map[string]interface{}{
					"gte": minSalary,
				},
			},
		})
	}

	if maxSalary, ok := filters["max_salary"].(float64); ok && maxSalary > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"salary_min": map[string]interface{}{
					"lte": maxSalary,
				},
			},
		})
	}

	if expLevel, ok := filters["experience_level"].(string); ok && expLevel != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"experience_level": expLevel,
			},
		})
	}

	if empType, ok := filters["employment_type"].(string); ok && empType != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"employment_type": empType,
			},
		})
	}

	if !includeInactive {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"is_active": true,
			},
		})
	}

	return query
}

func BuildJobAggregations() map[string]interface{} {
	return map[string]interface{}{
		"by_location": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "location.keyword",
				"size":  20,
			},
		},
		"by_skills": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "required_skills",
				"size":  30,
			},
		},
		"by_experience_level": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "experience_level",
				"size":  10,
			},
		},
		"by_employment_type": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "employment_type",
				"size":  10,
			},
		},
		"salary_range": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "salary_max",
			},
		},
	}
}

func (i *CandidateIndexer) IndexCandidate(ctx context.Context, candidate *models.EmployeeProfile, github *models.GithubProfile) error {
	doc := &CandidateDocument{
		ID:                candidate.UserID,
		UserID:            candidate.UserID,
		FullName:          candidate.FullName,
		Headline:          candidate.Headline,
		Bio:               candidate.Bio,
		Skills:            candidate.Skills,
		ExperienceLevel:   candidate.ExperienceLevel,
		YearsOfExperience: candidate.YearsOfExperience,
		Location:          candidate.Location,
		IsRemote:          candidate.IsRemote,
		IsHybrid:          candidate.IsHybrid,
		GithubUsername:    candidate.GithubUsername,
		IsAvailable:       candidate.IsAvailable,
		CreatedAt:         candidate.CreatedAt,
	}

	if github != nil {
		doc.GitHubActivity = GitHubActivityDoc{
			PublicRepos:   github.PublicRepos,
			TotalCommits:  github.TotalCommits,
			Followers:     github.Followers,
			TopLanguages:  extractTopLanguages(github.RepoLanguages),
			ActivityScore: github.ActivityScore,
		}
	}

	return i.esClient.IndexCandidate(ctx, doc)
}

func (i *CandidateIndexer) IndexAllCandidates(ctx context.Context) error {
	log.Println("Indexing all candidates to Elasticsearch...")

	page := 1
	limit := 100
	totalIndexed := 0

	for {
		candidates, total, err := i.userRepo.ListActiveEmployees(ctx, page, limit)
		if err != nil {
			return err
		}

		docs := make([]BulkDoc, 0, len(candidates))
		for _, candidate := range candidates {
			github, _ := i.userRepo.GetGithubProfileByEmployeeID(ctx, candidate.UserID)

			doc := &CandidateDocument{
				ID:                candidate.UserID,
				UserID:            candidate.UserID,
				FullName:          candidate.FullName,
				Headline:          candidate.Headline,
				Bio:               candidate.Bio,
				Skills:            candidate.Skills,
				ExperienceLevel:   candidate.ExperienceLevel,
				YearsOfExperience: candidate.YearsOfExperience,
				Location:          candidate.Location,
				IsRemote:          candidate.IsRemote,
				IsHybrid:          candidate.IsHybrid,
				GithubUsername:    candidate.GithubUsername,
				IsAvailable:       candidate.IsAvailable,
				CreatedAt:         candidate.CreatedAt,
			}

			if github != nil {
				doc.GitHubActivity = GitHubActivityDoc{
					PublicRepos:   github.PublicRepos,
					TotalCommits:  github.TotalCommits,
					Followers:     github.Followers,
					TopLanguages:  extractTopLanguages(github.RepoLanguages),
					ActivityScore: github.ActivityScore,
				}
			}

			docs = append(docs, BulkDoc{ID: candidate.UserID, Body: doc})
		}

		if len(docs) > 0 {
			if err := i.esClient.BulkIndex(ctx, "candidates", docs); err != nil {
				log.Printf("Failed to bulk index %d candidates: %v", len(docs), err)
			} else {
				totalIndexed += len(docs)
			}
		}

		if int64(page*limit) >= total {
			break
		}
		page++
	}

	log.Printf("Indexed %d candidates to Elasticsearch", totalIndexed)
	return nil
}

func (i *CandidateIndexer) DeleteCandidate(ctx context.Context, candidateID string) error {
	return i.esClient.DeleteDocument(ctx, "candidates", candidateID)
}

func (i *CompanyIndexer) IndexCompany(ctx context.Context, company *models.EmployerProfile) error {
	doc := &CompanyDocument{
		ID:          company.UserID,
		Name:        company.CompanyName,
		Description: company.CompanyDescription,
		Industry:    company.Industry,
		Size:        company.CompanySize,
		Location:    company.Location,
		Website:     company.CompanyWebsite,
		IsVerified:  company.VerificationStatus == "verified",
		TotalJobs:   company.TotalJobsPosted,
		CreatedAt:   company.CreatedAt,
	}

	return i.esClient.IndexCompany(ctx, doc)
}

func (i *CompanyIndexer) IndexAllCompanies(ctx context.Context) error {
	log.Println("Indexing all companies to Elasticsearch...")

	page := 1
	limit := 100
	totalIndexed := 0

	for {
		companies, total, err := i.userRepo.ListEmployers(ctx, page, limit)
		if err != nil {
			return err
		}

		docs := make([]BulkDoc, 0, len(companies))
		for _, company := range companies {
			doc := &CompanyDocument{
				ID:          company.UserID,
				Name:        company.CompanyName,
				Description: company.CompanyDescription,
				Industry:    company.Industry,
				Size:        company.CompanySize,
				Location:    company.Location,
				Website:     company.CompanyWebsite,
				IsVerified:  company.VerificationStatus == "verified",
				TotalJobs:   company.TotalJobsPosted,
				CreatedAt:   company.CreatedAt,
			}
			docs = append(docs, BulkDoc{ID: company.UserID, Body: doc})
		}

		if len(docs) > 0 {
			if err := i.esClient.BulkIndex(ctx, "companies", docs); err != nil {
				log.Printf("Failed to bulk index %d companies: %v", len(docs), err)
			} else {
				totalIndexed += len(docs)
			}
		}

		if int64(page*limit) >= total {
			break
		}
		page++
	}

	log.Printf("Indexed %d companies to Elasticsearch", totalIndexed)
	return nil
}

func (i *CompanyIndexer) DeleteCompany(ctx context.Context, companyID string) error {
	return i.esClient.DeleteDocument(ctx, "companies", companyID)
}

func extractTopLanguages(langs interface{}) []string {
	m, ok := langs.(map[string]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
