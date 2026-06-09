package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type UserRepository interface {
	// User operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	VerifyEmail(ctx context.Context, userID string) error
	UpdateAvatar(ctx context.Context, userID string, avatarURL string) error
	
	// Employee profile operations
	CreateEmployeeProfile(ctx context.Context, profile *models.EmployeeProfile) error
	GetEmployeeProfile(ctx context.Context, userID string) (*models.EmployeeProfile, error)
	GetEmployeeProfileByGithubUsername(ctx context.Context, username string) (*models.EmployeeProfile, error)
	UpdateEmployeeProfile(ctx context.Context, userID string, updates map[string]interface{}) error
	DeleteEmployeeProfile(ctx context.Context, userID string) error
	
	// Employer profile operations
	CreateEmployerProfile(ctx context.Context, profile *models.EmployerProfile) error
	GetEmployerProfile(ctx context.Context, userID string) (*models.EmployerProfile, error)
	GetEmployerProfileByCompanyName(ctx context.Context, companyName string) (*models.EmployerProfile, error)
	UpdateEmployerProfile(ctx context.Context, userID string, updates map[string]interface{}) error
	DeleteEmployerProfile(ctx context.Context, userID string) error
	UpdateEmployerVerification(ctx context.Context, userID string, status string, verifiedBy string) error
	IncrementJobPostedCount(ctx context.Context, userID string) error
	IncrementHiresCount(ctx context.Context, userID string) error

	// Methods for matching service
	ListActiveEmployees(ctx context.Context, page, limit int) ([]*models.EmployeeProfile, int64, error)
	ListEmployeesBySkills(ctx context.Context, skills []string, page, limit int) ([]*models.EmployeeProfile, int64, error)
	GetEmployeeWithGitHub(ctx context.Context, userID string) (*models.EmployeeProfile, *models.GithubProfile, error)

	// Employer batch operations
	ListEmployers(ctx context.Context, page, limit int) ([]*models.EmployerProfile, int64, error)
	GetEmployerProfilesBatch(ctx context.Context, userIDs []string, dest map[string]*models.EmployerProfile) error

	// GitHub profile
	GetGithubProfileByEmployeeID(ctx context.Context, employeeID string) (*models.GithubProfile, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// User operations

func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	// Soft delete by setting is_active to false
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).Error
}

func (r *UserRepositoryImpl) VerifyEmail(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_verified": true,
			"updated_at":  time.Now(),
		}).Error
}

func (r *UserRepositoryImpl) UpdateAvatar(ctx context.Context, userID string, avatarURL string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"avatar_url": avatarURL,
			"updated_at": time.Now(),
		}).Error
}

// Employee profile operations

func (r *UserRepositoryImpl) CreateEmployeeProfile(ctx context.Context, profile *models.EmployeeProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *UserRepositoryImpl) GetEmployeeProfile(ctx context.Context, userID string) (*models.EmployeeProfile, error) {
	var profile models.EmployeeProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepositoryImpl) GetEmployeeProfileByGithubUsername(ctx context.Context, username string) (*models.EmployeeProfile, error) {
	var profile models.EmployeeProfile
	err := r.db.WithContext(ctx).
		Where("github_username = ?", username).
		Preload("User").
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepositoryImpl) UpdateEmployeeProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&models.EmployeeProfile{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *UserRepositoryImpl) DeleteEmployeeProfile(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.EmployeeProfile{}).Error
}

// Employer profile operations

func (r *UserRepositoryImpl) CreateEmployerProfile(ctx context.Context, profile *models.EmployerProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *UserRepositoryImpl) GetEmployerProfile(ctx context.Context, userID string) (*models.EmployerProfile, error) {
	var profile models.EmployerProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepositoryImpl) GetEmployerProfileByCompanyName(ctx context.Context, companyName string) (*models.EmployerProfile, error) {
	var profile models.EmployerProfile
	err := r.db.WithContext(ctx).
		Where("company_name = ?", companyName).
		Preload("User").
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *UserRepositoryImpl) UpdateEmployerProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *UserRepositoryImpl) DeleteEmployerProfile(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.EmployerProfile{}).Error
}

func (r *UserRepositoryImpl) UpdateEmployerVerification(ctx context.Context, userID string, status string, verifiedBy string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"verification_status": status,
		"verified_at":         now,
		"verified_by":         verifiedBy,
		"updated_at":          now,
	}
	return r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *UserRepositoryImpl) IncrementJobPostedCount(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("total_jobs_posted", gorm.Expr("total_jobs_posted + 1")).Error
}

func (r *UserRepositoryImpl) IncrementHiresCount(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("total_hires", gorm.Expr("total_hires + 1")).Error
}



func (r *UserRepositoryImpl) ListActiveEmployees(ctx context.Context, page, limit int) ([]*models.EmployeeProfile, int64, error) {
	var employees []*models.EmployeeProfile
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&models.EmployeeProfile{}).
		Joins("JOIN users ON users.id = employee_profiles.user_id").
		Where("users.is_active = ? AND users.is_verified = ?", true, true)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("GithubProfile").
		Offset(offset).
		Limit(limit).
		Find(&employees).Error
	
	return employees, total, err
}

func (r *UserRepositoryImpl) ListEmployeesBySkills(ctx context.Context, skills []string, page, limit int) ([]*models.EmployeeProfile, int64, error) {
	var employees []*models.EmployeeProfile
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&models.EmployeeProfile{}).
		Joins("JOIN users ON users.id = employee_profiles.user_id").
		Where("users.is_active = ? AND users.is_verified = ?", true, true)
	
	if len(skills) > 0 {
		for _, skill := range skills {
			query = query.Where("? = ANY(skills)", skill)
		}
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("GithubProfile").
		Offset(offset).
		Limit(limit).
		Find(&employees).Error
	
	return employees, total, err
}

func (r *UserRepositoryImpl) GetEmployeeWithGitHub(ctx context.Context, userID string) (*models.EmployeeProfile, *models.GithubProfile, error) {
	var employee models.EmployeeProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		First(&employee).Error
	if err != nil {
		return nil, nil, err
	}
	
	var github models.GithubProfile
	err = r.db.WithContext(ctx).
		Where("employee_id = ?", userID).
		First(&github).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &employee, nil, nil
	}
	
	return &employee, &github, nil
}

func (r *UserRepositoryImpl) ListEmployers(ctx context.Context, page, limit int) ([]*models.EmployerProfile, int64, error) {
	var employers []*models.EmployerProfile
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.EmployerProfile{}).
		Joins("JOIN users ON users.id = employer_profiles.user_id").
		Where("users.is_active = ? AND users.is_verified = ?", true, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Offset(offset).
		Limit(limit).
		Find(&employers).Error

	return employers, total, err
}

func (r *UserRepositoryImpl) GetEmployerProfilesBatch(ctx context.Context, userIDs []string, dest map[string]*models.EmployerProfile) error {
	var employers []*models.EmployerProfile
	err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Find(&employers).Error
	if err != nil {
		return err
	}
	for _, emp := range employers {
		dest[emp.UserID] = emp
	}
	return nil
}

func (r *UserRepositoryImpl) GetGithubProfileByEmployeeID(ctx context.Context, employeeID string) (*models.GithubProfile, error) {
	var github models.GithubProfile
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&github).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &github, nil
}