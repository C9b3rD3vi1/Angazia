package services

import (
	"context"
	"fmt"
	"time"

	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type TalentPoolService interface {
	// Pool management
	CreatePool(ctx context.Context, employerID string, req *models.CreateTalentPoolRequest) (*models.TalentPool, error)
	GetPool(ctx context.Context, poolID, employerID string) (*models.TalentPool, error)
	UpdatePool(ctx context.Context, poolID, employerID string, req *models.UpdateTalentPoolRequest) (*models.TalentPool, error)
	DeletePool(ctx context.Context, poolID, employerID string) error
	ListPools(ctx context.Context, employerID string, page, limit int) (*models.TalentPoolListResponse, error)
	
	// Candidate management
	AddCandidate(ctx context.Context, poolID, employerID string, req *models.AddCandidateRequest) (*models.TalentPoolCandidate, error)
	UpdateCandidate(ctx context.Context, candidateID, poolID, employerID string, req *models.UpdateCandidateRequest) (*models.TalentPoolCandidate, error)
	RemoveCandidate(ctx context.Context, candidateID, poolID, employerID string) error
	ListCandidates(ctx context.Context, poolID, employerID string, filters repository.CandidateFilters, page, limit int) (*models.CandidateListResponse, error)
	
	// Bulk operations
	BulkAddCandidates(ctx context.Context, poolID, employerID string, employeeIDs []string, matchScores map[string]int) error
	BulkUpdateStatus(ctx context.Context, poolID, employerID string, candidateIDs []string, status string) error
	BulkRemoveCandidates(ctx context.Context, poolID, employerID string, candidateIDs []string) error
	
	// Actions
	MarkCandidateContacted(ctx context.Context, candidateID, poolID, employerID string) error
	MarkCandidateHired(ctx context.Context, candidateID, poolID, employerID string) error
	
	// Candidate pool lookup
	GetCandidatePools(ctx context.Context, employerID, employeeID string) ([]*models.TalentPool, error)

	// Statistics
	GetPoolStats(ctx context.Context, poolID, employerID string) (*models.TalentPoolStats, error)
	GetEmployerStats(ctx context.Context, employerID string) (map[string]int, error)
	
	// Search
	SearchCandidates(ctx context.Context, poolID, employerID, query string, page, limit int) (*models.CandidateListResponse, error)
}

type TalentPoolServiceImpl struct {
	cfg          *config.Config
	talentRepo   repository.TalentPoolRepository
	userRepo     repository.UserRepository
	jobRepo      repository.JobRepository
	matchingSvc  MatchingService
}

func NewTalentPoolService(
	cfg *config.Config,
	talentRepo repository.TalentPoolRepository,
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	matchingSvc MatchingService,
) TalentPoolService {
	return &TalentPoolServiceImpl{
		cfg:         cfg,
		talentRepo:  talentRepo,
		userRepo:    userRepo,
		jobRepo:     jobRepo,
		matchingSvc: matchingSvc,
	}
}

func (s *TalentPoolServiceImpl) CreatePool(ctx context.Context, employerID string, req *models.CreateTalentPoolRequest) (*models.TalentPool, error) {
	pool := &models.TalentPool{
		EmployerID:  employerID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}
	
	if req.Color != "" {
		pool.Color = req.Color
	}
	if req.Icon != "" {
		pool.Icon = req.Icon
	}
	
	if err := s.talentRepo.CreatePool(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create talent pool: %w", err)
	}
	
	return pool, nil
}

func (s *TalentPoolServiceImpl) GetPool(ctx context.Context, poolID, employerID string) (*models.TalentPool, error) {
	return s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID)
}

func (s *TalentPoolServiceImpl) UpdatePool(ctx context.Context, poolID, employerID string, req *models.UpdateTalentPoolRequest) (*models.TalentPool, error) {
	pool, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		pool.Name = *req.Name
	}
	if req.Description != nil {
		pool.Description = *req.Description
	}
	if req.Color != nil {
		pool.Color = *req.Color
	}
	if req.Icon != nil {
		pool.Icon = *req.Icon
	}
	if req.IsActive != nil {
		pool.IsActive = *req.IsActive
	}
	
	if err := s.talentRepo.UpdatePool(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to update talent pool: %w", err)
	}
	
	return pool, nil
}

func (s *TalentPoolServiceImpl) DeletePool(ctx context.Context, poolID, employerID string) error {
	return s.talentRepo.DeletePool(ctx, poolID, employerID)
}

func (s *TalentPoolServiceImpl) ListPools(ctx context.Context, employerID string, page, limit int) (*models.TalentPoolListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 50 {
		limit = 50
	}
	
	pools, total, err := s.talentRepo.ListPoolsByEmployer(ctx, employerID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &models.TalentPoolListResponse{
		Pools:      pools,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *TalentPoolServiceImpl) AddCandidate(ctx context.Context, poolID, employerID string, req *models.AddCandidateRequest) (*models.TalentPoolCandidate, error) {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return nil, fmt.Errorf("pool not found")
	}
	
	// Check if candidate already exists in pool
	existing, _ := s.talentRepo.GetCandidateByPoolAndEmployee(ctx, poolID, req.EmployeeID)
	if existing != nil {
		return nil, fmt.Errorf("candidate already exists in this pool")
	}
	
	// Calculate match score using matching service
	matchScore := req.MatchScore
	if matchScore == 0 {
		// Get all active jobs from this employer to calculate average match score
		jobs, _, err := s.jobRepo.ListByEmployer(ctx, employerID, 1, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get employer jobs: %w", err)
		}
		
		// Get candidate profile
		employee, err := s.userRepo.GetEmployeeProfile(ctx, req.EmployeeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get candidate profile: %w", err)
		}
		
		// Calculate match score with each job and take average
		totalScore := 0
		jobCount := 0
		
		for _, job := range jobs {
			if !job.IsActive {
				continue
			}
			
			// Get match analysis from AI
			analysis, err := s.matchingSvc.GetDetailedMatchAnalysis(ctx, job.ID, req.EmployeeID)
			if err == nil && analysis != nil {
				totalScore += analysis.OverallScore
				jobCount++
			}
		}
		
		if jobCount > 0 {
			matchScore = totalScore / jobCount
		} else {
			// Fallback: calculate simple match score based on skills
			matchScore = s.calculateSimpleMatchScore(employee)
		}
		
		// Ensure score is between 0 and 100
		if matchScore < 0 {
			matchScore = 0
		}
		if matchScore > 100 {
			matchScore = 100
		}
	}
	
	candidate := &models.TalentPoolCandidate{
		TalentPoolID: poolID,
		EmployeeID:   req.EmployeeID,
		MatchScore:   matchScore,
		Notes:        req.Notes,
		Tags:         req.Tags,
		Status:       "active",
	}
	
	if err := s.talentRepo.AddCandidate(ctx, candidate); err != nil {
		return nil, fmt.Errorf("failed to add candidate: %w", err)
	}
	
	return candidate, nil
}

// calculateSimpleMatchScore provides a fallback match score when AI service is unavailable
func (s *TalentPoolServiceImpl) calculateSimpleMatchScore(employee *models.EmployeeProfile) int {
	score := 50 // Base score
	
	// Skills weight (30 points)
	if len(employee.Skills) >= 5 {
		score += 15
	} else if len(employee.Skills) >= 3 {
		score += 10
	} else if len(employee.Skills) >= 1 {
		score += 5
	}
	
	// Experience weight (25 points)
	if employee.YearsOfExperience >= 5 {
		score += 20
	} else if employee.YearsOfExperience >= 3 {
		score += 15
	} else if employee.YearsOfExperience >= 1 {
		score += 10
	} else if employee.YearsOfExperience > 0 {
		score += 5
	}
	
	// Profile completeness (20 points)
	if employee.FullName != "" {
		score += 5
	}
	if employee.Headline != "" {
		score += 5
	}
	if employee.Bio != "" && len(employee.Bio) > 50 {
		score += 5
	}
	if employee.Location != "" {
		score += 5
	}
	
	// GitHub connection (15 points)
	if employee.GithubConnected && employee.GithubUsername != "" {
		score += 15
	}
	
	// Availability (10 points)
	if employee.IsAvailable {
		score += 10
	}
	
	if score > 100 {
		score = 100
	}
	
	return score
}

func (s *TalentPoolServiceImpl) UpdateCandidate(ctx context.Context, candidateID, poolID, employerID string, req *models.UpdateCandidateRequest) (*models.TalentPoolCandidate, error) {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return nil, fmt.Errorf("pool not found")
	}
	
	candidate, err := s.talentRepo.GetCandidate(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	
	if candidate.TalentPoolID != poolID {
		return nil, fmt.Errorf("candidate not found in this pool")
	}
	
	if req.Notes != nil {
		candidate.Notes = *req.Notes
	}
	if req.Tags != nil {
		candidate.Tags = req.Tags
	}
	if req.Status != nil {
		candidate.Status = *req.Status
	}
	
	if err := s.talentRepo.UpdateCandidate(ctx, candidate); err != nil {
		return nil, fmt.Errorf("failed to update candidate: %w", err)
	}
	
	return candidate, nil
}

func (s *TalentPoolServiceImpl) RemoveCandidate(ctx context.Context, candidateID, poolID, employerID string) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	return s.talentRepo.RemoveCandidate(ctx, candidateID, poolID)
}

func (s *TalentPoolServiceImpl) ListCandidates(ctx context.Context, poolID, employerID string, filters repository.CandidateFilters, page, limit int) (*models.CandidateListResponse, error) {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return nil, fmt.Errorf("pool not found")
	}
	
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	candidates, total, err := s.talentRepo.ListCandidatesByPool(ctx, poolID, filters, page, limit)
	if err != nil {
		return nil, err
	}
	
	// Build response with details
	candidatesWithDetails := make([]*models.TalentPoolCandidateWithDetails, len(candidates))
	for i, c := range candidates {
		details := &models.TalentPoolCandidateWithDetails{
			TalentPoolCandidate: *c,
		}
		
		if c.Employee != nil {
			details.EmployeeName = c.Employee.FullName
			details.EmployeeHeadline = c.Employee.Headline
			details.EmployeeSkills = c.Employee.Skills
			details.EmployeeLocation = c.Employee.Location
			details.YearsExperience = c.Employee.YearsOfExperience
			details.GitHubUsername = c.Employee.GithubUsername
			
			if c.Employee.User != nil {
				details.EmployeeEmail = c.Employee.User.Email
			}
		}
		
		candidatesWithDetails[i] = details
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &models.CandidateListResponse{
		Candidates: candidatesWithDetails,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *TalentPoolServiceImpl) BulkAddCandidates(ctx context.Context, poolID, employerID string, employeeIDs []string, matchScores map[string]int) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	candidates := make([]*models.TalentPoolCandidate, 0, len(employeeIDs))
	for _, empID := range employeeIDs {
		// Check if already exists
		existing, _ := s.talentRepo.GetCandidateByPoolAndEmployee(ctx, poolID, empID)
		if existing != nil {
			continue
		}
		
		matchScore := 50
		if score, ok := matchScores[empID]; ok {
			matchScore = score
		} else {
			// Calculate match score using the same logic as AddCandidate
			jobs, _, err := s.jobRepo.ListByEmployer(ctx, employerID, 1, 100)
			if err == nil {
				employee, err := s.userRepo.GetEmployeeProfile(ctx, empID)
				if err == nil {
					totalScore := 0
					jobCount := 0
					
					for _, job := range jobs {
						if !job.IsActive {
							continue
						}
						
						analysis, err := s.matchingSvc.GetDetailedMatchAnalysis(ctx, job.ID, empID)
						if err == nil && analysis != nil {
							totalScore += analysis.OverallScore
							jobCount++
						}
					}
					
					if jobCount > 0 {
						matchScore = totalScore / jobCount
					} else {
						matchScore = s.calculateSimpleMatchScore(employee)
					}
				}
			}
		}
		
		candidates = append(candidates, &models.TalentPoolCandidate{
			TalentPoolID: poolID,
			EmployeeID:   empID,
			MatchScore:   matchScore,
			Status:       "active",
		})
	}
	
	if len(candidates) == 0 {
		return nil
	}
	
	return s.talentRepo.BulkAddCandidates(ctx, candidates)
}

func (s *TalentPoolServiceImpl) BulkUpdateStatus(ctx context.Context, poolID, employerID string, candidateIDs []string, status string) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	return s.talentRepo.BulkUpdateStatus(ctx, candidateIDs, status)
}

func (s *TalentPoolServiceImpl) BulkRemoveCandidates(ctx context.Context, poolID, employerID string, candidateIDs []string) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	return s.talentRepo.BulkRemoveCandidates(ctx, candidateIDs, poolID)
}

func (s *TalentPoolServiceImpl) MarkCandidateContacted(ctx context.Context, candidateID, poolID, employerID string) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	candidate, err := s.talentRepo.GetCandidate(ctx, candidateID)
	if err != nil {
		return err
	}
	
	if candidate.TalentPoolID != poolID {
		return fmt.Errorf("candidate not found in this pool")
	}
	
	now := time.Now()
	candidate.Status = "contacted"
	candidate.ContactedAt = &now
	
	return s.talentRepo.UpdateCandidate(ctx, candidate)
}

func (s *TalentPoolServiceImpl) MarkCandidateHired(ctx context.Context, candidateID, poolID, employerID string) error {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return fmt.Errorf("pool not found")
	}
	
	candidate, err := s.talentRepo.GetCandidate(ctx, candidateID)
	if err != nil {
		return err
	}
	
	if candidate.TalentPoolID != poolID {
		return fmt.Errorf("candidate not found in this pool")
	}
	
	now := time.Now()
	candidate.Status = "hired"
	candidate.HiredAt = &now
	
	return s.talentRepo.UpdateCandidate(ctx, candidate)
}

func (s *TalentPoolServiceImpl) GetPoolStats(ctx context.Context, poolID, employerID string) (*models.TalentPoolStats, error) {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return nil, fmt.Errorf("pool not found")
	}
	
	return s.talentRepo.GetPoolStats(ctx, poolID)
}

func (s *TalentPoolServiceImpl) GetEmployerStats(ctx context.Context, employerID string) (map[string]int, error) {
	return s.talentRepo.GetEmployerTalentStats(ctx, employerID)
}

func (s *TalentPoolServiceImpl) GetCandidatePools(ctx context.Context, employerID, employeeID string) ([]*models.TalentPool, error) {
	return s.talentRepo.GetCandidatePools(ctx, employerID, employeeID)
}

func (s *TalentPoolServiceImpl) SearchCandidates(ctx context.Context, poolID, employerID, query string, page, limit int) (*models.CandidateListResponse, error) {
	// Verify pool belongs to employer
	if _, err := s.talentRepo.GetPoolByEmployer(ctx, poolID, employerID); err != nil {
		return nil, fmt.Errorf("pool not found")
	}
	
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	candidates, total, err := s.talentRepo.SearchCandidatesInPool(ctx, poolID, query, page, limit)
	if err != nil {
		return nil, err
	}
	
	candidatesWithDetails := make([]*models.TalentPoolCandidateWithDetails, len(candidates))
	for i, c := range candidates {
		details := &models.TalentPoolCandidateWithDetails{
			TalentPoolCandidate: *c,
		}
		
		if c.Employee != nil {
			details.EmployeeName = c.Employee.FullName
			details.EmployeeHeadline = c.Employee.Headline
			details.EmployeeSkills = c.Employee.Skills
			details.EmployeeLocation = c.Employee.Location
			details.YearsExperience = c.Employee.YearsOfExperience
			details.GitHubUsername = c.Employee.GithubUsername
			
			if c.Employee.User != nil {
				details.EmployeeEmail = c.Employee.User.Email
			}
		}
		
		candidatesWithDetails[i] = details
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &models.CandidateListResponse{
		Candidates: candidatesWithDetails,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}