package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
//	"github.com/google/uuid"
	
	"github.com/Angazia/internal/models"
)

type GitHubRepository interface {
	CreateProfile(ctx context.Context, profile *models.GithubProfile) error
	GetProfileByEmployeeID(ctx context.Context, employeeID string) (*models.GithubProfile, error)
	GetProfileByGitHubID(ctx context.Context, githubID int) (*models.GithubProfile, error)
	GetProfileByUsername(ctx context.Context, username string) (*models.GithubProfile, error)
	UpdateProfile(ctx context.Context, profile *models.GithubProfile) error
	UpdateProfileStats(ctx context.Context, employeeID string, updates map[string]interface{}) error
	
	CreateToken(ctx context.Context, token *GitHubTokenDB) error
	GetTokenByEmployeeID(ctx context.Context, employeeID string) (*GitHubTokenDB, error)
	UpdateToken(ctx context.Context, employeeID string, updates map[string]interface{}) error
	DeleteToken(ctx context.Context, employeeID string) error
	
	SaveRepositories(ctx context.Context, repos []*models.GithubRepository) error
	GetRepositories(ctx context.Context, employeeID string, filters map[string]interface{}, page, limit int) ([]*models.GithubRepository, int64, error)
	DeleteRepositories(ctx context.Context, employeeID string) error
	
	SaveContributions(ctx context.Context, contributions []*models.GithubContribution) error
	GetContributions(ctx context.Context, employeeID string, since, until time.Time) ([]*models.GithubContribution, error)
	
	CreateSyncLog(ctx context.Context, log *models.GithubSyncLog) error
	GetLastSyncLog(ctx context.Context, employeeID string) (*models.GithubSyncLog, error)
}

type GitHubRepositoryImpl struct {
	db *gorm.DB
}

// GitHubTokenDB is the database model for OAuth tokens
type GitHubTokenDB struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID   string    `gorm:"type:uuid;not null;uniqueIndex"`
	AccessToken  string    `gorm:"not null"`
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scope        string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (GitHubTokenDB) TableName() string {
	return "github_tokens"
}

func NewGitHubRepository(db *gorm.DB) GitHubRepository {
	return &GitHubRepositoryImpl{db: db}
}

func (r *GitHubRepositoryImpl) CreateProfile(ctx context.Context, profile *models.GithubProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *GitHubRepositoryImpl) GetProfileByEmployeeID(ctx context.Context, employeeID string) (*models.GithubProfile, error) {
	var profile models.GithubProfile
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *GitHubRepositoryImpl) GetProfileByGitHubID(ctx context.Context, githubID int) (*models.GithubProfile, error) {
	var profile models.GithubProfile
	err := r.db.WithContext(ctx).
		Where("github_id = ?", githubID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *GitHubRepositoryImpl) GetProfileByUsername(ctx context.Context, username string) (*models.GithubProfile, error) {
	var profile models.GithubProfile
	err := r.db.WithContext(ctx).
		Where("github_username = ?", username).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *GitHubRepositoryImpl) UpdateProfile(ctx context.Context, profile *models.GithubProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *GitHubRepositoryImpl) UpdateProfileStats(ctx context.Context, employeeID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.GithubProfile{}).
		Where("employee_id = ?", employeeID).
		Updates(updates).Error
}

func (r *GitHubRepositoryImpl) CreateToken(ctx context.Context, token *GitHubTokenDB) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *GitHubRepositoryImpl) GetTokenByEmployeeID(ctx context.Context, employeeID string) (*GitHubTokenDB, error) {
	var token GitHubTokenDB
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *GitHubRepositoryImpl) UpdateToken(ctx context.Context, employeeID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&GitHubTokenDB{}).
		Where("employee_id = ?", employeeID).
		Updates(updates).Error
}

func (r *GitHubRepositoryImpl) DeleteToken(ctx context.Context, employeeID string) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Delete(&GitHubTokenDB{}).Error
}

func (r *GitHubRepositoryImpl) SaveRepositories(ctx context.Context, repos []*models.GithubRepository) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, repo := range repos {
			var existing models.GithubRepository
			err := tx.Where("employee_id = ? AND repo_id = ?", repo.EmployeeID, repo.RepoID).
				First(&existing).Error
			
			if err == nil {
				// Update existing
				repo.ID = existing.ID
				if err := tx.Save(repo).Error; err != nil {
					return err
				}
			} else {
				// Create new
				if err := tx.Create(repo).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *GitHubRepositoryImpl) GetRepositories(ctx context.Context, employeeID string, filters map[string]interface{}, page, limit int) ([]*models.GithubRepository, int64, error) {
	var repos []*models.GithubRepository
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.GithubRepository{}).
		Where("employee_id = ?", employeeID)
	
	// Apply filters
	if language, ok := filters["language"].(string); ok && language != "" {
		query = query.Where("language = ?", language)
	}
	
	if isFork, ok := filters["is_fork"].(bool); ok {
		query = query.Where("is_fork = ?", isFork)
	}
	
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply sorting
	if sortBy, ok := filters["sort_by"].(string); ok {
		sortOrder := "DESC"
		if order, ok := filters["sort_order"].(string); ok && order == "asc" {
			sortOrder = "ASC"
		}
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("stars DESC")
	}
	
	// Pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	
	return repos, total, nil
}

func (r *GitHubRepositoryImpl) DeleteRepositories(ctx context.Context, employeeID string) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Delete(&models.GithubRepository{}).Error
}

func (r *GitHubRepositoryImpl) SaveContributions(ctx context.Context, contributions []*models.GithubContribution) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, contrib := range contributions {
			var existing models.GithubContribution
			err := tx.Where("employee_id = ? AND date = ?", contrib.EmployeeID, contrib.Date).
				First(&existing).Error
			
			if err == nil {
				// Update existing
				existing.Count = contrib.Count
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else {
				// Create new
				if err := tx.Create(contrib).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *GitHubRepositoryImpl) GetContributions(ctx context.Context, employeeID string, since, until time.Time) ([]*models.GithubContribution, error) {
	var contributions []*models.GithubContribution
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, since, until).
		Order("date ASC").
		Find(&contributions).Error
	return contributions, err
}

func (r *GitHubRepositoryImpl) CreateSyncLog(ctx context.Context, log *models.GithubSyncLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GitHubRepositoryImpl) GetLastSyncLog(ctx context.Context, employeeID string) (*models.GithubSyncLog, error) {
	var log models.GithubSyncLog
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("synced_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}