package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type CompanyService interface {
	// Profile management
	GetCompanyProfile(ctx context.Context, companyID string) (*CompanyProfileResponse, error)
	UpdateCompanyProfile(ctx context.Context, companyID string, req *UpdateCompanyProfileRequest) (*models.EmployerProfile, error)
	UploadCompanyLogo(ctx context.Context, companyID string, file multipart.File, filename string) (string, error)
	
	// Verification
	SubmitVerification(ctx context.Context, companyID string, req *VerificationRequest) (*models.CompanyVerification, error)
	GetVerificationStatus(ctx context.Context, companyID string) (*models.CompanyVerification, error)
	
	// Trust badges
	GetCompanyBadges(ctx context.Context, companyID string) ([]*models.TrustBadge, error)
	
	// Reviews
	SubmitReview(ctx context.Context, companyID, userID string, req *SubmitReviewRequest) (*models.CompanyReview, error)
	GetCompanyReviews(ctx context.Context, companyID string, page, limit int) (*ReviewListResponse, error)
	GetReviewStats(ctx context.Context, companyID string) (*models.ReviewStats, error)
	MarkReviewHelpful(ctx context.Context, reviewID, userID string) error
	
	// Team management
	InviteTeamMember(ctx context.Context, companyID, inviterID string, req *InviteTeamMemberRequest) (*models.TeamInvitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) error
	GetTeamMembers(ctx context.Context, companyID string) ([]*TeamMember, error)
	RemoveTeamMember(ctx context.Context, companyID, memberID string) error
	UpdateTeamMemberRole(ctx context.Context, companyID, memberID, role string) error
	
	// Public company page
	GetPublicCompanyProfile(ctx context.Context, companyID string) (*PublicCompanyProfile, error)
	
	// Analytics
	GetCompanyAnalytics(ctx context.Context, companyID string, days int) (*CompanyAnalyticsResponse, error)
}

type CompanyProfileResponse struct {
	Profile      *models.EmployerProfile        `json:"profile"`
	Verification *models.CompanyVerification    `json:"verification,omitempty"`
	Badges       []*models.TrustBadge           `json:"badges"`
	Stats        *CompanyStats                  `json:"stats"`
}

type CompanyStats struct {
	TotalJobs      int     `json:"total_jobs"`
	ActiveJobs     int     `json:"active_jobs"`
	TotalHires     int     `json:"total_hires"`
	TotalReviews   int     `json:"total_reviews"`
	AverageRating  float64 `json:"average_rating"`
}

type UpdateCompanyProfileRequest struct {
	CompanyName                  string `json:"company_name"`
	CompanyWebsite               string `json:"company_website"`
	CompanyLinkedIn              string `json:"company_linkedin"`
	CompanyDescription           string `json:"company_description"`
	Industry                     string `json:"industry"`
	CompanySize                  string `json:"company_size"`
	Location                     string `json:"location"`
	PhoneNumber                  string `json:"phone_number"`
	EmailAddress                 string `json:"email_address"`
	BusinessRegistrationNumber   string `json:"business_registration_number"`
	TaxID                        string `json:"tax_id"`
}

type VerificationRequest struct {
	BusinessRegistrationNumber string   `json:"business_registration_number"`
	TaxID                      string   `json:"tax_id"`
	Documents                  []string `json:"documents"`
}

type SubmitReviewRequest struct {
	Rating           int    `json:"rating" validate:"required,min=1,max=5"`
	Title            string `json:"title" validate:"max=255"`
	Content          string `json:"content" validate:"required,min=10"`
	Pros             string `json:"pros"`
	Cons             string `json:"cons"`
	WouldRecommend   bool   `json:"would_recommend"`
	EmploymentStatus string `json:"employment_status" validate:"oneof=former current interviewed"`
}

type InviteTeamMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin recruiter viewer"`
}

type TeamMember struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	IsOwner   bool      `json:"is_owner"`
}

type ReviewListResponse struct {
	Reviews    []*models.CompanyReview `json:"reviews"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

type PublicCompanyProfile struct {
	ID          string               `json:"id"`
	CompanyName string               `json:"company_name"`
	CompanyLogo string               `json:"company_logo"`
	Description string               `json:"description"`
	Industry    string               `json:"industry"`
	CompanySize string               `json:"company_size"`
	Location    string               `json:"location"`
	Website     string               `json:"website"`
	LinkedIn    string               `json:"linkedin"`
	Badges      []*models.TrustBadge `json:"badges"`
	Stats       *CompanyStats        `json:"stats"`
	Rating      float64              `json:"rating"`
	ReviewCount int                  `json:"review_count"`
}

type CompanyAnalyticsResponse struct {
	ProfileViews         []DailyStat `json:"profile_views"`
	JobViews             []DailyStat `json:"job_views"`
	ApplicationsReceived []DailyStat `json:"applications_received"`
	TotalProfileViews    int         `json:"total_profile_views"`
	TotalJobViews        int         `json:"total_job_views"`
	TotalApplications    int         `json:"total_applications"`
}

type DailyStat struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

type CompanyServiceImpl struct {
	cfg          *config.Config
	companyRepo  repository.CompanyRepository
	employerRepo repository.UserRepository
	jobRepo      repository.JobRepository
	emailService EmailService
}

func NewCompanyService(
	cfg *config.Config,
	companyRepo repository.CompanyRepository,
	employerRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	emailService EmailService,
) CompanyService {
	return &CompanyServiceImpl{
		cfg:          cfg,
		companyRepo:  companyRepo,
		employerRepo: employerRepo,
		jobRepo:      jobRepo,
		emailService: emailService,
	}
}

func (s *CompanyServiceImpl) GetCompanyProfile(ctx context.Context, companyID string) (*CompanyProfileResponse, error) {
	profile, err := s.employerRepo.GetEmployerProfile(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company profile: %w", err)
	}
	
	verification, _ := s.companyRepo.GetVerification(ctx, companyID)
	badges, _ := s.companyRepo.GetBadges(ctx, companyID)
	
	// Get job stats
	jobs, total, err := s.jobRepo.ListByEmployer(ctx, companyID, 1, 100)
	if err != nil {
		jobs = []*models.Job{}
		total = 0
	}
	
	activeJobs := int64(0)
	for _, job := range jobs {
		if job.IsActive {
			activeJobs++
		}
	}
	
	reviewStats, _ := s.companyRepo.GetReviewStats(ctx, companyID)
	
	stats := &CompanyStats{
		TotalJobs:     int(total),
		ActiveJobs:    int(activeJobs),
		TotalHires:    profile.TotalHires,
		TotalReviews:  reviewStats.TotalReviews,
		AverageRating: reviewStats.AverageRating,
	}
	
	return &CompanyProfileResponse{
		Profile:      profile,
		Verification: verification,
		Badges:       badges,
		Stats:        stats,
	}, nil
}

func (s *CompanyServiceImpl) UpdateCompanyProfile(ctx context.Context, companyID string, req *UpdateCompanyProfileRequest) (*models.EmployerProfile, error) {
	updates := make(map[string]interface{})
	
	if req.CompanyName != "" {
		updates["company_name"] = req.CompanyName
	}
	if req.CompanyWebsite != "" {
		updates["company_website"] = req.CompanyWebsite
	}
	if req.CompanyLinkedIn != "" {
		updates["company_linkedin"] = req.CompanyLinkedIn
	}
	if req.CompanyDescription != "" {
		updates["company_description"] = req.CompanyDescription
	}
	if req.Industry != "" {
		updates["industry"] = req.Industry
	}
	if req.CompanySize != "" {
		updates["company_size"] = req.CompanySize
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.PhoneNumber != "" {
		updates["phone_number"] = req.PhoneNumber
	}
	if req.EmailAddress != "" {
		updates["contact_email"] = req.EmailAddress
	}
	
	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	
	if err := s.employerRepo.UpdateEmployerProfile(ctx, companyID, updates); err != nil {
		return nil, fmt.Errorf("failed to update company profile: %w", err)
	}

	if req.BusinessRegistrationNumber != "" || req.TaxID != "" {
		if err := s.companyRepo.UpsertVerificationDetails(ctx, companyID, req.BusinessRegistrationNumber, req.TaxID); err != nil {
			return nil, fmt.Errorf("failed to update verification details: %w", err)
		}
	}

	return s.employerRepo.GetEmployerProfile(ctx, companyID)
}

func (s *CompanyServiceImpl) UploadCompanyLogo(ctx context.Context, companyID string, file multipart.File, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return "", fmt.Errorf("invalid file type. Allowed: JPG, PNG, WEBP")
	}

	newFilename := fmt.Sprintf("company_%s_%d%s", companyID, time.Now().Unix(), ext)
	uploadDir := filepath.Join(s.cfg.UploadDir, "companies")
	filePath := filepath.Join(uploadDir, newFilename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	logoURL := fmt.Sprintf("%s/uploads/companies/%s", s.cfg.AppURL, newFilename)

	updates := map[string]interface{}{
		"company_logo": logoURL,
	}

	if err := s.employerRepo.UpdateEmployerProfile(ctx, companyID, updates); err != nil {
		return "", fmt.Errorf("failed to update company logo: %w", err)
	}

	return logoURL, nil
}

func (s *CompanyServiceImpl) SubmitVerification(ctx context.Context, companyID string, req *VerificationRequest) (*models.CompanyVerification, error) {
	existing, _ := s.companyRepo.GetVerification(ctx, companyID)
	if existing != nil {
		if existing.Status == "pending" {
			return nil, fmt.Errorf("verification already pending")
		}
		if existing.Status == "approved" {
			return nil, fmt.Errorf("company already verified")
		}
	}

	businessReg := req.BusinessRegistrationNumber
	taxID := req.TaxID
	var documents models.JSONArray

	if existing != nil {
		if businessReg == "" {
			businessReg = existing.BusinessRegistrationNumber
		}
		if taxID == "" {
			taxID = existing.TaxID
		}
		if len(req.Documents) == 0 {
			documents = existing.Documents
		}
	}

	if len(req.Documents) > 0 {
		documents = make(models.JSONArray, len(req.Documents))
		for i, d := range req.Documents {
			documents[i] = d
		}
	}
	
	verification := &models.CompanyVerification{
		CompanyID:                  companyID,
		BusinessRegistrationNumber: businessReg,
		TaxID:                      taxID,
		Documents:                  documents,
		Status:                     "pending",
	}
	
	if err := s.companyRepo.CreateVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("failed to submit verification: %w", err)
	}
	
	// Send notification to admin using email service with template
	go s.sendVerificationNotification(verification)
	
	return verification, nil
}

func (s *CompanyServiceImpl) GetVerificationStatus(ctx context.Context, companyID string) (*models.CompanyVerification, error) {
	return s.companyRepo.GetVerification(ctx, companyID)
}

func (s *CompanyServiceImpl) GetCompanyBadges(ctx context.Context, companyID string) ([]*models.TrustBadge, error) {
	return s.companyRepo.GetBadges(ctx, companyID)
}

func (s *CompanyServiceImpl) SubmitReview(ctx context.Context, companyID, userID string, req *SubmitReviewRequest) (*models.CompanyReview, error) {
	hasReviewed, err := s.companyRepo.HasUserReviewed(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if hasReviewed {
		return nil, fmt.Errorf("you have already reviewed this company")
	}
	
	review := &models.CompanyReview{
		CompanyID:        companyID,
		ReviewerID:       userID,
		Rating:           req.Rating,
		Title:            req.Title,
		Content:          req.Content,
		Pros:             req.Pros,
		Cons:             req.Cons,
		WouldRecommend:   req.WouldRecommend,
		EmploymentStatus: req.EmploymentStatus,
	}
	
	if err := s.companyRepo.CreateReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to submit review: %w", err)
	}
	
	return review, nil
}

func (s *CompanyServiceImpl) GetCompanyReviews(ctx context.Context, companyID string, page, limit int) (*ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	
	reviews, total, err := s.companyRepo.GetReviewsByCompany(ctx, companyID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &ReviewListResponse{
		Reviews:    reviews,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *CompanyServiceImpl) GetReviewStats(ctx context.Context, companyID string) (*models.ReviewStats, error) {
	return s.companyRepo.GetReviewStats(ctx, companyID)
}

func (s *CompanyServiceImpl) MarkReviewHelpful(ctx context.Context, reviewID, userID string) error {
	return s.companyRepo.IncrementHelpfulCount(ctx, reviewID)
}

func (s *CompanyServiceImpl) InviteTeamMember(ctx context.Context, companyID, inviterID string, req *InviteTeamMemberRequest) (*models.TeamInvitation, error) {
	existingUser, _ := s.employerRepo.GetByEmail(ctx, req.Email)
	
	invitation := &models.TeamInvitation{
		CompanyID: companyID,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: inviterID,
	}
	
	if err := s.companyRepo.CreateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}
	
	// Send invitation email using email service with template
	go s.sendTeamInvitationEmail(invitation, existingUser != nil)
	
	return invitation, nil
}

func (s *CompanyServiceImpl) AcceptInvitation(ctx context.Context, token, userID string) error {
	invitation, err := s.companyRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired invitation")
	}
	
	updates := map[string]interface{}{
		"company_id": invitation.CompanyID,
	}
	if err := s.employerRepo.UpdateEmployerProfile(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}
	
	return s.companyRepo.UpdateInvitationStatus(ctx, token, "accepted")
}

func (s *CompanyServiceImpl) GetTeamMembers(ctx context.Context, companyID string) ([]*TeamMember, error) {
	profile, err := s.employerRepo.GetEmployerProfile(ctx, companyID)
	if err != nil {
		return nil, err
	}
	
	members := []*TeamMember{
		{
			ID:       profile.UserID,
			Email:    profile.User.Email,
			FullName: profile.CompanyName,
			Role:     "owner",
			JoinedAt: profile.CreatedAt,
			IsOwner:  true,
		},
	}
	
	invitations, _, err := s.companyRepo.GetInvitationsByCompany(ctx, companyID, 1, 100)
	if err == nil {
		for _, inv := range invitations {
			if inv.Status == "accepted" {
				member, _ := s.employerRepo.GetByEmail(ctx, inv.Email)
				if member != nil {
					members = append(members, &TeamMember{
						ID:       member.ID,
						Email:    member.Email,
						FullName: member.Email,
						Role:     inv.Role,
						JoinedAt: *inv.AcceptedAt,
						IsOwner:  false,
					})
				}
			}
		}
	}
	
	return members, nil
}

func (s *CompanyServiceImpl) RemoveTeamMember(ctx context.Context, companyID, memberID string) error {
	profile, err := s.employerRepo.GetEmployerProfile(ctx, companyID)
	if err != nil {
		return err
	}
	if profile.UserID == memberID {
		return fmt.Errorf("cannot remove company owner")
	}
	
	updates := map[string]interface{}{
		"company_id": nil,
	}
	return s.employerRepo.UpdateEmployerProfile(ctx, memberID, updates)
}

func (s *CompanyServiceImpl) UpdateTeamMemberRole(ctx context.Context, companyID, memberID, role string) error {
	profile, err := s.employerRepo.GetEmployerProfile(ctx, companyID)
	if err != nil {
		return err
	}
	if profile.UserID == memberID {
		return fmt.Errorf("cannot change owner role")
	}
	
	invitations, _, err := s.companyRepo.GetInvitationsByCompany(ctx, companyID, 1, 100)
	if err != nil {
		return err
	}
	
	for _, inv := range invitations {
		if inv.Status == "accepted" && inv.Email == profile.User.Email {
			inv.Role = role
			break
		}
	}
	
	return nil
}

func (s *CompanyServiceImpl) GetPublicCompanyProfile(ctx context.Context, companyID string) (*PublicCompanyProfile, error) {
	profile, err := s.employerRepo.GetEmployerProfile(ctx, companyID)
	if err != nil {
		return nil, err
	}
	
	badges, _ := s.companyRepo.GetBadges(ctx, companyID)
	
	jobs, total, err := s.jobRepo.ListByEmployer(ctx, companyID, 1, 100)
	if err != nil {
		jobs = []*models.Job{}
		total = 0
	}
	
	activeJobs := int64(0)
	for _, job := range jobs {
		if job.IsActive {
			activeJobs++
		}
	}
	
	reviewStats, _ := s.companyRepo.GetReviewStats(ctx, companyID)
	
	stats := &CompanyStats{
		TotalJobs:     int(total),
		ActiveJobs:    int(activeJobs),
		TotalHires:    profile.TotalHires,
		TotalReviews:  reviewStats.TotalReviews,
		AverageRating: reviewStats.AverageRating,
	}
	
	go s.companyRepo.IncrementProfileViews(ctx, companyID)
	
	return &PublicCompanyProfile{
		ID:          profile.UserID,
		CompanyName: profile.CompanyName,
		CompanyLogo: profile.CompanyLogo,
		Description: profile.CompanyDescription,
		Industry:    profile.Industry,
		CompanySize: profile.CompanySize,
		Location:    profile.Location,
		Website:     profile.CompanyWebsite,
		LinkedIn:    profile.CompanyLinkedIn,
		Badges:      badges,
		Stats:       stats,
		Rating:      reviewStats.AverageRating,
		ReviewCount: reviewStats.TotalReviews,
	}, nil
}

func (s *CompanyServiceImpl) GetCompanyAnalytics(ctx context.Context, companyID string, days int) (*CompanyAnalyticsResponse, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	
	analytics, err := s.companyRepo.GetAnalytics(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	response := &CompanyAnalyticsResponse{
		ProfileViews:         []DailyStat{},
		JobViews:             []DailyStat{},
		ApplicationsReceived: []DailyStat{},
	}
	
	analyticsMap := make(map[string]*models.CompanyAnalytics)
	for _, a := range analytics {
		analyticsMap[a.Date.Format("2006-01-02")] = a
	}
	
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		var profileViews, jobViews, applications int
		
		if a, ok := analyticsMap[dateStr]; ok {
			profileViews = a.ProfileViews
			jobViews = a.JobViews
			applications = a.ApplicationsReceived
			
			response.TotalProfileViews += profileViews
			response.TotalJobViews += jobViews
			response.TotalApplications += applications
		}
		
		response.ProfileViews = append(response.ProfileViews, DailyStat{Date: dateStr, Value: profileViews})
		response.JobViews = append(response.JobViews, DailyStat{Date: dateStr, Value: jobViews})
		response.ApplicationsReceived = append(response.ApplicationsReceived, DailyStat{Date: dateStr, Value: applications})
	}
	
	return response, nil
}

// ========== EMAIL METHODS USING TEMPLATES ==========

func (s *CompanyServiceImpl) sendVerificationNotification(verification *models.CompanyVerification) {
	if s.emailService == nil {
		fmt.Printf("Verification submitted for company %s\n", verification.CompanyID)
		return
	}
	
	// Get admin email from config
	adminEmail := s.cfg.AdminEmail
	if adminEmail == "" {
		fmt.Printf("No admin email configured. Verification pending for company %s\n", verification.CompanyID)
		return
	}
	
	// Use email service to send admin notification
	// The email service will use HTML templates
	s.emailService.SendAdminVerificationNotification(
		adminEmail,
		verification.CompanyID,
		verification.BusinessRegistrationNumber,
		verification.TaxID,
		verification.ID,
	)
}

func (s *CompanyServiceImpl) sendTeamInvitationEmail(invitation *models.TeamInvitation, hasAccount bool) {
	if s.emailService == nil {
		fmt.Printf("Invitation sent to %s for company %s\n", invitation.Email, invitation.CompanyID)
		return
	}
	
	if hasAccount {
		s.emailService.SendTeamInvitationExistingUser(
			invitation.Email,
			invitation.CompanyID,
			invitation.Role,
			invitation.Token,
		)
	} else {
		s.emailService.SendTeamInvitationNewUser(
			invitation.Email,
			invitation.CompanyID,
			invitation.Role,
			invitation.Token,
		)
	}
}