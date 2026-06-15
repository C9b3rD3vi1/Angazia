package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type CompanyRepository interface {
	// Verification
	CreateVerification(ctx context.Context, verification *models.CompanyVerification) error
	GetVerification(ctx context.Context, companyID string) (*models.CompanyVerification, error)
	UpdateVerificationStatus(ctx context.Context, companyID, status, rejectionReason, verifiedBy string) error
	UpsertVerificationDetails(ctx context.Context, companyID, businessReg, taxID string) error
	GetPendingVerifications(ctx context.Context, page, limit int) ([]*models.CompanyVerification, int64, error)
	
	// Trust Badges
	AddBadge(ctx context.Context, badge *models.TrustBadge) error
	GetBadges(ctx context.Context, companyID string) ([]*models.TrustBadge, error)
	RemoveBadge(ctx context.Context, badgeID string) error
	HasBadge(ctx context.Context, companyID, badgeType string) (bool, error)
	
	// Reviews
	CreateReview(ctx context.Context, review *models.CompanyReview) error
	GetReview(ctx context.Context, id string) (*models.CompanyReview, error)
	GetReviewsByCompany(ctx context.Context, companyID string, page, limit int) ([]*models.CompanyReview, int64, error)
	GetReviewStats(ctx context.Context, companyID string) (*models.ReviewStats, error)
	IncrementHelpfulCount(ctx context.Context, reviewID string) error
	HasUserReviewed(ctx context.Context, companyID, userID string) (bool, error)
	
	// Team Invitations
	CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error
	GetInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error)
	GetInvitationsByCompany(ctx context.Context, companyID string, page, limit int) ([]*models.TeamInvitation, int64, error)
	UpdateInvitationStatus(ctx context.Context, token, status string) error
	UpdateInvitationRole(ctx context.Context, invitationID, role string) error
	AcceptInvitation(ctx context.Context, token, userID string) error
	CancelInvitation(ctx context.Context, invitationID, companyID string) error
	GetTeamMembers(ctx context.Context, companyID string) ([]*models.User, error)
	
	// Analytics
	TrackAnalytics(ctx context.Context, analytics *models.CompanyAnalytics) error
	GetAnalytics(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.CompanyAnalytics, error)
	IncrementProfileViews(ctx context.Context, companyID string) error

	// Reviews (additional methods)
	GetReviewsByUser(ctx context.Context, userID string, page, limit int) ([]*models.CompanyReview, int64, error)
	GetReviewByID(ctx context.Context, reviewID string) (*models.CompanyReview, error)
	DeleteReview(ctx context.Context, reviewID string) error
	GetReviewsByRating(ctx context.Context, companyID string, rating int, page, limit int) ([]*models.CompanyReview, int64, error)
}

type CompanyRepositoryImpl struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &CompanyRepositoryImpl{db: db}
}

func (r *CompanyRepositoryImpl) CreateVerification(ctx context.Context, verification *models.CompanyVerification) error {
	verification.ID = uuid.New().String()
	verification.SubmittedAt = time.Now()
	return r.db.WithContext(ctx).Create(verification).Error
}

func (r *CompanyRepositoryImpl) UpsertVerificationDetails(ctx context.Context, companyID, businessReg, taxID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.CompanyVerification{}).
		Where("company_id = ?", companyID).
		Updates(map[string]interface{}{
			"business_registration_number": businessReg,
			"tax_id":                       taxID,
			"updated_at":                   now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return r.db.WithContext(ctx).Create(&models.CompanyVerification{
			ID:                         uuid.New().String(),
			CompanyID:                  companyID,
			BusinessRegistrationNumber: businessReg,
			TaxID:                      taxID,
			SubmittedAt:                now,
		}).Error
	}

	return nil
}

func (r *CompanyRepositoryImpl) GetVerification(ctx context.Context, companyID string) (*models.CompanyVerification, error) {
	var verification models.CompanyVerification
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *CompanyRepositoryImpl) UpdateVerificationStatus(ctx context.Context, companyID, status, rejectionReason, verifiedBy string) error {
	updates := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}
	
	if status == "approved" {
		updates["verified_by"] = verifiedBy
		updates["verified_at"] = time.Now()
	}
	
	if status == "rejected" && rejectionReason != "" {
		updates["rejection_reason"] = rejectionReason
	}
	
	return r.db.WithContext(ctx).
		Model(&models.CompanyVerification{}).
		Where("company_id = ?", companyID).
		Updates(updates).Error
}

func (r *CompanyRepositoryImpl) GetPendingVerifications(ctx context.Context, page, limit int) ([]*models.CompanyVerification, int64, error) {
	var verifications []*models.CompanyVerification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.CompanyVerification{}).
		Where("status = ?", "pending").
		Preload("Company").
		Preload("Company.User")
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("submitted_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&verifications).Error
	
	return verifications, total, err
}

func (r *CompanyRepositoryImpl) AddBadge(ctx context.Context, badge *models.TrustBadge) error {
	badge.ID = uuid.New().String()
	badge.AwardedAt = time.Now()
	return r.db.WithContext(ctx).Create(badge).Error
}

func (r *CompanyRepositoryImpl) GetBadges(ctx context.Context, companyID string) ([]*models.TrustBadge, error) {
	var badges []*models.TrustBadge
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Find(&badges).Error
	return badges, err
}

func (r *CompanyRepositoryImpl) RemoveBadge(ctx context.Context, badgeID string) error {
	return r.db.WithContext(ctx).
		Model(&models.TrustBadge{}).
		Where("id = ?", badgeID).
		Update("is_active", false).Error
}

func (r *CompanyRepositoryImpl) HasBadge(ctx context.Context, companyID, badgeType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TrustBadge{}).
		Where("company_id = ? AND badge_type = ? AND is_active = ?", companyID, badgeType, true).
		Count(&count).Error
	return count > 0, err
}

func (r *CompanyRepositoryImpl) CreateReview(ctx context.Context, review *models.CompanyReview) error {
	review.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *CompanyRepositoryImpl) GetReview(ctx context.Context, id string) (*models.CompanyReview, error) {
	var review models.CompanyReview
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Reviewer").
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *CompanyRepositoryImpl) GetReviewsByCompany(ctx context.Context, companyID string, page, limit int) ([]*models.CompanyReview, int64, error) {
	var reviews []*models.CompanyReview
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.CompanyReview{}).
		Where("company_id = ?", companyID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Reviewer").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error
	
	return reviews, total, err
}

func (r *CompanyRepositoryImpl) GetReviewStats(ctx context.Context, companyID string) (*models.ReviewStats, error) {
	stats := &models.ReviewStats{
		RatingDistribution: make(map[int]int),
	}
	
	var results []struct {
		Rating int
		Count  int
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.CompanyReview{}).
		Select("rating, COUNT(*) as count").
		Where("company_id = ?", companyID).
		Group("rating").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	var totalRating int
	for _, r := range results {
		stats.RatingDistribution[r.Rating] = int(r.Count)
		stats.TotalReviews += int(r.Count)
		totalRating += r.Rating * int(r.Count)
	}
	
	if stats.TotalReviews > 0 {
		stats.AverageRating = float64(totalRating) / float64(stats.TotalReviews)
	}
	
	// Calculate recommendation rate
	var recommendCount int64
	r.db.WithContext(ctx).
		Model(&models.CompanyReview{}).
		Where("company_id = ? AND would_recommend = ?", companyID, true).
		Count(&recommendCount)
	
	stats.WouldRecommend = int(recommendCount)
	stats.WouldNotRecommend = stats.TotalReviews - int(recommendCount)
	
	if stats.TotalReviews > 0 {
		stats.RecommendationRate = (float64(recommendCount) / float64(stats.TotalReviews)) * 100
	}
	
	return stats, nil
}

func (r *CompanyRepositoryImpl) IncrementHelpfulCount(ctx context.Context, reviewID string) error {
	return r.db.WithContext(ctx).
		Model(&models.CompanyReview{}).
		Where("id = ?", reviewID).
		UpdateColumn("helpful_count", gorm.Expr("helpful_count + 1")).Error
}

func (r *CompanyRepositoryImpl) HasUserReviewed(ctx context.Context, companyID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CompanyReview{}).
		Where("company_id = ? AND reviewer_id = ?", companyID, userID).
		Count(&count).Error
	return count > 0, err
}


func (r *CompanyRepositoryImpl) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	invitation.ID = uuid.New().String()
	invitation.Token = uuid.New().String()
	invitation.ExpiresAt = time.Now().AddDate(0, 0, 7) // 7 days
	return r.db.WithContext(ctx).Create(invitation).Error
}

func (r *CompanyRepositoryImpl) GetInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	var invitation models.TeamInvitation
	err := r.db.WithContext(ctx).
		Where("token = ? AND status = ? AND expires_at > ?", token, "pending", time.Now()).
		Preload("Company").
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *CompanyRepositoryImpl) GetInvitationsByCompany(ctx context.Context, companyID string, page, limit int) ([]*models.TeamInvitation, int64, error) {
	var invitations []*models.TeamInvitation
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.TeamInvitation{}).
		Where("company_id = ?", companyID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Inviter").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invitations).Error
	
	return invitations, total, err
}

func (r *CompanyRepositoryImpl) UpdateInvitationStatus(ctx context.Context, token, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.TeamInvitation{}).
		Where("token = ?", token).
		Update("status", status).Error
}

func (r *CompanyRepositoryImpl) UpdateInvitationRole(ctx context.Context, invitationID, role string) error {
	return r.db.WithContext(ctx).
		Model(&models.TeamInvitation{}).
		Where("id = ?", invitationID).
		Update("role", role).Error
}

func (r *CompanyRepositoryImpl) AcceptInvitation(ctx context.Context, token, userID string) error {
	var invitation models.TeamInvitation
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&invitation).Error; err != nil {
		return err
	}
	
	// Update user's employer profile with company ID
	if err := r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Where("user_id = ?", userID).
		Update("company_id", invitation.CompanyID).Error; err != nil {
		return err
	}
	
	// Mark invitation as accepted
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.TeamInvitation{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"status":      "accepted",
			"accepted_at": &now,
		}).Error
}

func (r *CompanyRepositoryImpl) CancelInvitation(ctx context.Context, invitationID, companyID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND company_id = ?", invitationID, companyID).
		Delete(&models.TeamInvitation{}).Error
}

func (r *CompanyRepositoryImpl) GetTeamMembers(ctx context.Context, companyID string) ([]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN employer_profiles ON employer_profiles.user_id = users.id").
		Where("employer_profiles.company_id = ?", companyID).
		Find(&users).Error
	return users, err
}

func (r *CompanyRepositoryImpl) TrackAnalytics(ctx context.Context, analytics *models.CompanyAnalytics) error {
	analytics.ID = uuid.New().String()
	analytics.Date = time.Now().Truncate(24 * time.Hour)
	return r.db.WithContext(ctx).Create(analytics).Error
}

func (r *CompanyRepositoryImpl) GetAnalytics(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.CompanyAnalytics, error) {
	var analytics []*models.CompanyAnalytics
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND date BETWEEN ? AND ?", companyID, startDate, endDate).
		Order("date ASC").
		Find(&analytics).Error
	return analytics, err
}

func (r *CompanyRepositoryImpl) IncrementProfileViews(ctx context.Context, companyID string) error {
	today := time.Now().Truncate(24 * time.Hour)
	
	// Upsert analytics for today
	var analytics models.CompanyAnalytics
	result := r.db.WithContext(ctx).
		Where("company_id = ? AND date = ?", companyID, today).
		First(&analytics)
	
	if result.Error == nil {
		return r.db.WithContext(ctx).
			Model(&analytics).
			UpdateColumn("profile_views", gorm.Expr("profile_views + 1")).Error
	}
	
	// Create new record
	analytics = models.CompanyAnalytics{
		CompanyID:    companyID,
		Date:         today,
		ProfileViews: 1,
	}
	return r.TrackAnalytics(ctx, &analytics)
}


func (r *CompanyRepositoryImpl) GetReviewsByUser(ctx context.Context, userID string, page, limit int) ([]*models.CompanyReview, int64, error) {
	var reviews []*models.CompanyReview
	var total int64

	query := r.db.WithContext(ctx).Model(&models.CompanyReview{}).
		Where("reviewer_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Company").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	return reviews, total, err
}

func (r *CompanyRepositoryImpl) GetReviewByID(ctx context.Context, reviewID string) (*models.CompanyReview, error) {
	var review models.CompanyReview
	err := r.db.WithContext(ctx).
		Where("id = ?", reviewID).
		Preload("Reviewer").
		Preload("Company").
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *CompanyRepositoryImpl) DeleteReview(ctx context.Context, reviewID string) error {
	return r.db.WithContext(ctx).Delete(&models.CompanyReview{}, "id = ?", reviewID).Error
}

func (r *CompanyRepositoryImpl) GetReviewsByRating(ctx context.Context, companyID string, rating int, page, limit int) ([]*models.CompanyReview, int64, error) {
	var reviews []*models.CompanyReview
	var total int64

	query := r.db.WithContext(ctx).Model(&models.CompanyReview{}).
		Where("company_id = ? AND rating = ?", companyID, rating)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Reviewer").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	return reviews, total, err
}
