package models

import (
	"time"
	"encoding/json"
)

// GithubProfile stores GitHub user data
type GithubProfile struct {
	ID           string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID   string     `json:"employee_id" gorm:"type:uuid;not null;uniqueIndex"`
	GithubID     int        `json:"github_id" gorm:"uniqueIndex"` // GitHub's user ID
	GithubUsername string   `json:"github_username" gorm:"size:255;not null;index"`
	GithubEmail   string    `json:"github_email,omitempty" gorm:"size:255"`
	GithubAvatar  string    `json:"github_avatar,omitempty" gorm:"size:512"`
	GithubURL     string    `json:"github_url" gorm:"size:512"`
	GithubBio     string    `json:"github_bio,omitempty" gorm:"type:text"`
	GithubCompany string    `json:"github_company,omitempty" gorm:"size:255"`
	GithubLocation string   `json:"github_location,omitempty" gorm:"size:255"`
	GithubBlog    string    `json:"github_blog,omitempty" gorm:"size:512"`
	GithubJoined  time.Time `json:"github_joined"`
	
	// GitHub statistics
	PublicRepos   int       `json:"public_repos" gorm:"default:0"`
	PublicGists   int       `json:"public_gists" gorm:"default:0"`
	Followers     int       `json:"followers" gorm:"default:0"`
	Following     int       `json:"following" gorm:"default:0"`
	TotalCommits  int       `json:"total_commits" gorm:"default:0"` // Last year
	TotalPRs      int       `json:"total_prs" gorm:"default:0"`     // Last year
	TotalIssues   int       `json:"total_issues" gorm:"default:0"`   // Last year
	ContributionStreak int  `json:"contribution_streak" gorm:"default:0"`
	
	// Quality metrics
	RepoLanguages    JSONMap  `json:"repo_languages" gorm:"type:jsonb"` // {"Python": 45, "JavaScript": 30, "Go": 25}
	TopRepositories  JSONArray `json:"top_repositories" gorm:"type:jsonb"`
	CommitFrequency  string   `json:"commit_frequency" gorm:"size:50"` // daily, weekly, monthly, sporadic
	AccountAgeDays   int      `json:"account_age_days"`
	IsVerified       bool     `json:"is_verified" gorm:"default:false"` // GitHub verified email
	IsOrgMember      bool     `json:"is_org_member" gorm:"default:false"`
	
	// Activity scores (1-100)
	ActivityScore    int      `json:"activity_score" gorm:"default:0"`
	QualityScore     int      `json:"quality_score" gorm:"default:0"`
	OverallScore     int      `json:"overall_score" gorm:"default:0"`
	
	LastSyncedAt     time.Time `json:"last_synced_at"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Employee         *EmployeeProfile `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
}

// GithubContribution represents daily contributions
type GithubContribution struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID string    `json:"employee_id" gorm:"type:uuid;not null;index"`
	Date       time.Time `json:"date" gorm:"index"`
	Count      int       `json:"count" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// GithubRepository represents a GitHub repo (detailed)
type GithubRepository struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID    string    `json:"employee_id" gorm:"type:uuid;not null;index"`
	RepoID        int       `json:"repo_id" gorm:"index"` // GitHub repo ID
	Name          string    `json:"name" gorm:"size:255;not null"`
	FullName      string    `json:"full_name" gorm:"size:512"`
	Description   string    `json:"description" gorm:"type:text"`
	IsPrivate     bool      `json:"is_private" gorm:"default:false"`
	IsFork        bool      `json:"is_fork" gorm:"default:false"`
	Stars         int       `json:"stars" gorm:"default:0"`
	Forks         int       `json:"forks" gorm:"default:0"`
	Watchers      int       `json:"watchers" gorm:"default:0"`
	OpenIssues    int       `json:"open_issues" gorm:"default:0"`
	Language      string    `json:"language" gorm:"size:100"`
	SizeKB        int       `json:"size_kb"`
	CreatedAt     time.Time `json:"created_at"`
	PushedAt      time.Time `json:"pushed_at"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	LastFetched   time.Time `json:"last_fetched"`
	
	// Quality metrics
	HasReadme     bool      `json:"has_readme" gorm:"default:false"`
	HasWiki       bool      `json:"has_wiki" gorm:"default:false"`
	HasProjects   bool      `json:"has_projects" gorm:"default:false"`
	License       string    `json:"license,omitempty" gorm:"size:100"`
}

// GithubSyncLog tracks GitHub sync history
type GithubSyncLog struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID  string    `json:"employee_id" gorm:"type:uuid;not null;index"`
	SyncType    string    `json:"sync_type" gorm:"size:50"` // full, incremental
	Status      string    `json:"status" gorm:"size:50"` // success, failed, partial
	CommitsFetched int    `json:"commits_fetched" gorm:"default:0"`
	ReposFetched int      `json:"repos_fetched" gorm:"default:0"`
	ErrorMessage string    `json:"error_message,omitempty" gorm:"type:text"`
	DurationMs  int       `json:"duration_ms"`
	SyncedAt    time.Time `json:"synced_at" gorm:"autoCreateTime;index"`
}

// Helper structs for JSON fields
type JSONMap map[string]interface{}
type JSONArray []interface{}

// Scan/Value for JSONMap (for GORM)
func (j JSONMap) Value() (interface{}, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Scan/Value for JSONArray
func (j JSONArray) Value() (interface{}, error) {
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArray, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Helper methods for GithubProfile
func (g *GithubProfile) IsActiveContributor() bool {
	return g.ActivityScore >= 60
}

func (g *GithubProfile) GetTopLanguages(limit int) []string {
	if g.RepoLanguages == nil {
		return []string{}
	}
	
	type langCount struct {
		Name  string
		Count int
	}
	
	var langs []langCount
	for name, count := range g.RepoLanguages {
		if c, ok := count.(float64); ok {
			langs = append(langs, langCount{Name: name, Count: int(c)})
		} else if c, ok := count.(int); ok {
			langs = append(langs, langCount{Name: name, Count: c})
		}
	}
	
	// Sort by count descending
	for i := 0; i < len(langs)-1; i++ {
		for j := i + 1; j < len(langs); j++ {
			if langs[j].Count > langs[i].Count {
				langs[i], langs[j] = langs[j], langs[i]
			}
		}
	}
	
	result := make([]string, 0, limit)
	for i := 0; i < len(langs) && i < limit; i++ {
		result = append(result, langs[i].Name)
	}
	return result
}

func (g *GithubProfile) GetScoreCategory() string {
	if g.OverallScore >= 80 {
		return "Exceptional"
	} else if g.OverallScore >= 60 {
		return "Active Contributor"
	} else if g.OverallScore >= 40 {
		return "Casual Developer"
	} else {
		return "Getting Started"
	}
}

func (g *GithubProfile) TableName() string {
	return "github_profiles"
}

func (GithubContribution) TableName() string {
	return "github_contributions"
}

func (GithubRepository) TableName() string {
	return "github_repositories"
}

func (GithubSyncLog) TableName() string {
	return "github_sync_logs"
}
