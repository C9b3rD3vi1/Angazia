package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/github"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type GitHubService interface {
	// OAuth flows
	GetAuthURL(state, redirectTo string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)

	// User management
	HandleGitHubLogin(ctx context.Context, token *oauth2.Token, existingUserID string) (*GitHubAuthResult, error)
	ConnectGitHubAccount(ctx context.Context, userID string, token *oauth2.Token) error
	DisconnectGitHubAccount(ctx context.Context, userID string) error

	// Data sync
	SyncGitHubData(ctx context.Context, userID string) error
	GetGitHubProfile(ctx context.Context, userID string) (*models.GithubProfile, error)
	GetGitHubRepos(ctx context.Context, userID string, filters map[string]interface{}, page, limit int) ([]*models.GithubRepository, int64, error)
	GetGitHubContributions(ctx context.Context, userID string, days int) (*ContributionsResult, error)

	// Token management
	GetValidAccessToken(ctx context.Context, userID string) (string, error)

	// Webhook handling
	ProcessWebhook(ctx context.Context, eventType, deliveryID string, payload map[string]interface{}) error
}

type GitHubAuthResult struct {
	UserID       string
	Email        string
	IsNewUser    bool
	GitHubUser   *github.GitHubUser
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ContributionsResult struct {
	Contributions []*models.GithubContribution
	TotalCommits  int
	ActiveDays    int
	CurrentStreak int
	LongestStreak int
	ActivityLevel string
}

type GitHubServiceImpl struct {
	cfg          *config.Config
	oauthConfig  *oauth2.Config
	githubRepo   repository.GitHubRepository
	userRepo     repository.UserRepository
	githubClient *github.Client
}

func NewGitHubService(
	cfg *config.Config,
	githubRepo repository.GitHubRepository,
	userRepo repository.UserRepository,
) GitHubService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  cfg.GithubRedirectURL,
		Scopes: []string{
			"read:user",
			"user:email",
			"repo",
			"read:org",
			"read:repo_hook",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	return &GitHubServiceImpl{
		cfg:         cfg,
		oauthConfig: oauthConfig,
		githubRepo:  githubRepo,
		userRepo:    userRepo,
	}
}

func (s *GitHubServiceImpl) GetAuthURL(state, redirectTo string) string {
	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	}
	return s.oauthConfig.AuthCodeURL(state, opts...)
}

func (s *GitHubServiceImpl) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(ctx, code)
}

func (s *GitHubServiceImpl) HandleGitHubLogin(ctx context.Context, token *oauth2.Token, existingUserID string) (*GitHubAuthResult, error) {
	// Fetch GitHub user data
	githubClient := github.NewClient(token.AccessToken)
	githubUser, err := githubClient.GetAuthenticatedUser()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub user: %w", err)
	}

	// Get primary email
	email, err := githubClient.GetPrimaryEmail()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub email: %w", err)
	}

	result := &GitHubAuthResult{
		GitHubUser:   githubUser,
		Email:        email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}

	// If user is already logged in, just connect the account
	if existingUserID != "" {
		if err := s.connectGitHubToUser(ctx, existingUserID, githubUser, email, token); err != nil {
			return nil, err
		}
		result.UserID = existingUserID
		result.IsNewUser = false
		return result, nil
	}

	// Check if user already exists with this email
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// User exists, connect GitHub
		if err := s.connectGitHubToUser(ctx, existingUser.ID, githubUser, email, token); err != nil {
			return nil, err
		}
		result.UserID = existingUser.ID
		result.IsNewUser = false
		return result, nil
	}

	// Create new user
	userID, err := s.createUserFromGitHub(ctx, githubUser, email, token)
	if err != nil {
		return nil, err
	}

	result.UserID = userID
	result.IsNewUser = true
	return result, nil
}

func (s *GitHubServiceImpl) ConnectGitHubAccount(ctx context.Context, userID string, token *oauth2.Token) error {
	githubClient := github.NewClient(token.AccessToken)
	githubUser, err := githubClient.GetAuthenticatedUser()
	if err != nil {
		return fmt.Errorf("failed to fetch GitHub user: %w", err)
	}

	email, err := githubClient.GetPrimaryEmail()
	if err != nil {
		return fmt.Errorf("failed to fetch GitHub email: %w", err)
	}

	return s.connectGitHubToUser(ctx, userID, githubUser, email, token)
}

func (s *GitHubServiceImpl) DisconnectGitHubAccount(ctx context.Context, userID string) error {
	// Update employee profile
	if err := s.userRepo.UpdateEmployeeProfile(ctx, userID, map[string]interface{}{
		"github_connected": false,
		"github_username":  nil,
		"last_github_sync": nil,
	}); err != nil {
		return fmt.Errorf("failed to update employee profile: %w", err)
	}

	// Delete tokens
	if err := s.githubRepo.DeleteToken(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete GitHub token: %w", err)
	}

	return nil
}

func (s *GitHubServiceImpl) SyncGitHubData(ctx context.Context, userID string) error {
	startTime := time.Now()

	// Get valid access token
	accessToken, err := s.GetValidAccessToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	githubClient := github.NewClient(accessToken)

	// Create sync log
	syncLog := &models.GithubSyncLog{
		ID:         uuid.New().String(),
		EmployeeID: userID,
		SyncType:   "full",
		Status:     "processing",
		SyncedAt:   time.Now(),
	}

	// Fetch repositories
	repos, err := githubClient.GetRepositories()
	if err != nil {
		syncLog.Status = "failed"
		syncLog.ErrorMessage = err.Error()
		s.githubRepo.CreateSyncLog(ctx, syncLog)
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Convert to models
	modelRepos := make([]*models.GithubRepository, len(repos))
	for i, repo := range repos {
		modelRepos[i] = &models.GithubRepository{
			ID:          uuid.New().String(),
			EmployeeID:  userID,
			RepoID:      repo.ID,
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			IsPrivate:   repo.Private,
			IsFork:      repo.Fork,
			Stars:       repo.Stargazers,
			Forks:       repo.Forks,
			Watchers:    repo.Watchers,
			OpenIssues:  repo.OpenIssues,
			Language:    repo.Language,
			SizeKB:      repo.Size,
			CreatedAt:   repo.CreatedAt,
			PushedAt:    repo.PushedAt,
			HasWiki:     repo.HasWiki,
			HasProjects: repo.HasProjects,
			LastFetched: time.Now(),
		}
		if repo.License != nil {
			modelRepos[i].License = repo.License.Name
		}
	}

	// Save repositories
	if err := s.githubRepo.SaveRepositories(ctx, modelRepos); err != nil {
		syncLog.Status = "partial"
		syncLog.ErrorMessage = err.Error()
		s.githubRepo.CreateSyncLog(ctx, syncLog)
		return fmt.Errorf("failed to save repositories: %w", err)
	}

	// Fetch contributions (last 365 days)
	contributions, err := githubClient.GetContributions(365)
	if err != nil {
		syncLog.Status = "partial"
		syncLog.ErrorMessage = err.Error()
		s.githubRepo.CreateSyncLog(ctx, syncLog)
	} else if len(contributions) > 0 {
		// Save contributions
		modelContribs := make([]*models.GithubContribution, len(contributions))
		for i, contrib := range contributions {
			modelContribs[i] = &models.GithubContribution{
				ID:         uuid.New().String(),
				EmployeeID: userID,
				Date:       contrib.Date,
				Count:      contrib.Count,
			}
		}

		if err := s.githubRepo.SaveContributions(ctx, modelContribs); err != nil {
			syncLog.Status = "partial"
			syncLog.ErrorMessage = err.Error()
		} else {
			// Calculate and update scores
			s.updateGitHubScores(ctx, userID, modelRepos, modelContribs)
		}
	}

	// Update profile stats
	updates := map[string]interface{}{
		"public_repos":   len(repos),
		"last_synced_at": time.Now(),
	}
	s.githubRepo.UpdateProfileStats(ctx, userID, updates)

	// Update sync log
	syncLog.Status = "success"
	syncLog.ReposFetched = len(repos)
	syncLog.CommitsFetched = len(contributions)
	syncLog.DurationMs = int(time.Since(startTime).Milliseconds())
	s.githubRepo.CreateSyncLog(ctx, syncLog)

	return nil
}

func (s *GitHubServiceImpl) GetGitHubProfile(ctx context.Context, userID string) (*models.GithubProfile, error) {
	return s.githubRepo.GetProfileByEmployeeID(ctx, userID)
}

func (s *GitHubServiceImpl) GetGitHubRepos(ctx context.Context, userID string, filters map[string]interface{}, page, limit int) ([]*models.GithubRepository, int64, error) {
	return s.githubRepo.GetRepositories(ctx, userID, filters, page, limit)
}

func (s *GitHubServiceImpl) GetGitHubContributions(ctx context.Context, userID string, days int) (*ContributionsResult, error) {
	if days > 730 {
		days = 730
	}
	if days < 7 {
		days = 7
	}

	since := time.Now().AddDate(0, 0, -days)
	until := time.Now()

	contributions, err := s.githubRepo.GetContributions(ctx, userID, since, until)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	result := &ContributionsResult{
		Contributions: contributions,
	}

	contributionMap := make(map[string]int)
	for _, c := range contributions {
		dateStr := c.Date.Format("2006-01-02")
		contributionMap[dateStr] = c.Count
		if c.Count > 0 {
			result.TotalCommits += c.Count
			result.ActiveDays++
		}
	}

	// Calculate streaks
	currentDate := time.Now()
	var currentStreak int
	for i := 0; i < days; i++ {
		dateStr := currentDate.Format("2006-01-02")
		if count, exists := contributionMap[dateStr]; exists && count > 0 {
			currentStreak++
			if currentStreak > result.LongestStreak {
				result.LongestStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
		currentDate = currentDate.AddDate(0, 0, -1)
	}
	result.CurrentStreak = currentStreak

	// Calculate activity level
	percentage := (float64(result.ActiveDays) / float64(days)) * 100
	switch {
	case percentage >= 75:
		result.ActivityLevel = "Very Active 🔥"
	case percentage >= 50:
		result.ActivityLevel = "Active 💪"
	case percentage >= 25:
		result.ActivityLevel = "Moderately Active 📈"
	case percentage >= 10:
		result.ActivityLevel = "Occasionally Active 📊"
	default:
		result.ActivityLevel = "Getting Started 🌱"
	}

	return result, nil
}

func (s *GitHubServiceImpl) GetValidAccessToken(ctx context.Context, userID string) (string, error) {
	token, err := s.githubRepo.GetTokenByEmployeeID(ctx, userID)
	if err != nil {
		return "", err
	}

	// Decrypt access token using existing utils
	accessToken, err := utils.DecryptString(token.AccessToken)
	if err != nil {
		return "", err
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now().Add(5 * time.Minute)) {
		// Refresh token
		refreshToken, err := utils.DecryptString(token.RefreshToken)
		if err != nil {
			return "", err
		}

		newToken, err := s.refreshAccessToken(ctx, refreshToken)
		if err != nil {
			return "", err
		}

		// Update stored token
		encryptedAccess, _ := utils.EncryptString(newToken.AccessToken)
		encryptedRefresh, _ := utils.EncryptString(newToken.RefreshToken)

		updates := map[string]interface{}{
			"access_token":  encryptedAccess,
			"refresh_token": encryptedRefresh,
			"expires_at":    newToken.Expiry,
		}
		s.githubRepo.UpdateToken(ctx, userID, updates)

		return newToken.AccessToken, nil
	}

	return accessToken, nil
}

func (s *GitHubServiceImpl) ProcessWebhook(ctx context.Context, eventType, deliveryID string, payload map[string]interface{}) error {
	switch eventType {
	case "push":
		return s.handlePushWebhook(ctx, payload)
	case "star":
		return s.handleStarWebhook(ctx, payload)
	default:
		return nil
	}
}

// Private helper methods

func (s *GitHubServiceImpl) connectGitHubToUser(ctx context.Context, userID string, githubUser *github.GitHubUser, email string, token *oauth2.Token) error {
	// Check if GitHub account already connected to another user
	existingProfile, err := s.githubRepo.GetProfileByGitHubID(ctx, githubUser.ID)
	if err == nil && existingProfile.EmployeeID != userID {
		return fmt.Errorf("GitHub account already connected to another user")
	}

	// Update employee profile
	updates := map[string]interface{}{
		"github_connected": true,
		"github_username":  githubUser.Login,
		"last_github_sync": time.Now(),
	}

	profile, _ := s.userRepo.GetEmployeeProfile(ctx, userID)
	if profile != nil {
		if profile.FullName == "" && githubUser.Name != "" {
			updates["full_name"] = githubUser.Name
		}
		if profile.Bio == "" && githubUser.Bio != "" {
			updates["bio"] = truncateString(githubUser.Bio, 500)
		}
		if profile.Location == "" && githubUser.Location != "" {
			updates["location"] = githubUser.Location
		}
	}

	if err := s.userRepo.UpdateEmployeeProfile(ctx, userID, updates); err != nil {
		return err
	}

	// Store OAuth token with encryption using existing utils
	encryptedAccess, _ := utils.EncryptString(token.AccessToken)
	encryptedRefresh, _ := utils.EncryptString(token.RefreshToken)

	tokenDB := &repository.GitHubTokenDB{
		EmployeeID:   userID,
		AccessToken:  encryptedAccess,
		RefreshToken: encryptedRefresh,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        strings.Join(s.oauthConfig.Scopes, ","),
	}

	if err := s.githubRepo.CreateToken(ctx, tokenDB); err != nil {
		// Check if token exists, update if so
		if err := s.githubRepo.UpdateToken(ctx, userID, map[string]interface{}{
			"access_token":  encryptedAccess,
			"refresh_token": encryptedRefresh,
			"expires_at":    token.Expiry,
		}); err != nil {
			return err
		}
	}

	// Create or update GitHub profile
	githubProfile := s.buildGithubProfile(userID, githubUser, email)
	if err := s.githubRepo.UpdateProfile(ctx, githubProfile); err != nil {
		if err := s.githubRepo.CreateProfile(ctx, githubProfile); err != nil {
			return err
		}
	}

	// Start background sync
	go s.SyncGitHubData(context.Background(), userID)

	return nil
}

func (s *GitHubServiceImpl) createUserFromGitHub(ctx context.Context, githubUser *github.GitHubUser, email string, token *oauth2.Token) (string, error) {
	// Generate random password
	randomPassword := generateRandomPassword()
	hashedPassword, err := utils.HashPassword(randomPassword)
	if err != nil {
		return "", err
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         models.RoleEmployee,
		IsVerified:   true,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	// Create employee profile
	employeeProfile := &models.EmployeeProfile{
		UserID:          user.ID,
		FullName:        githubUser.Name,
		Headline:        "Software Developer",
		Bio:             truncateString(githubUser.Bio, 500),
		Location:        githubUser.Location,
		GithubConnected: true,
		GithubUsername:  &githubUser.Login,
		IsVisible:       true,
		IsAvailable:     true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if employeeProfile.FullName == "" {
		employeeProfile.FullName = githubUser.Login
	}

	if err := s.userRepo.CreateEmployeeProfile(ctx, employeeProfile); err != nil {
		s.userRepo.Delete(ctx, user.ID)
		return "", err
	}

	// Store token and GitHub profile
	if err := s.connectGitHubToUser(ctx, user.ID, githubUser, email, token); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (s *GitHubServiceImpl) buildGithubProfile(employeeID string, githubUser *github.GitHubUser, email string) *models.GithubProfile {
	profile := &models.GithubProfile{
		ID:             uuid.New().String(),
		EmployeeID:     employeeID,
		GithubID:       githubUser.ID,
		GithubUsername: githubUser.Login,
		GithubEmail:    email,
		GithubAvatar:   githubUser.AvatarURL,
		GithubURL:      githubUser.HTMLURL,
		GithubBio:      truncateString(githubUser.Bio, 500),
		GithubCompany:  githubUser.Company,
		GithubLocation: githubUser.Location,
		GithubBlog:     githubUser.Blog,
		PublicRepos:    githubUser.PublicRepos,
		Followers:      githubUser.Followers,
		Following:      githubUser.Following,
		LastSyncedAt:   time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Parse GitHub join date
	if githubUser.CreatedAt != nil {
		profile.GithubJoined = *githubUser.CreatedAt
		profile.AccountAgeDays = int(time.Since(*githubUser.CreatedAt).Hours() / 24)
	}

	// Calculate initial scores
	profile.ActivityScore = s.calculateInitialActivityScore(profile)
	profile.QualityScore = s.calculateInitialQualityScore(profile)
	profile.OverallScore = (profile.ActivityScore + profile.QualityScore) / 2
	profile.IsVerified = githubUser.Hireable

	return profile
}

func (s *GitHubServiceImpl) updateGitHubScores(ctx context.Context, userID string, repos []*models.GithubRepository, contributions []*models.GithubContribution) {
	// Calculate activity score based on contributions
	var totalCommits int
	var activeDays int
	for _, c := range contributions {
		if c.Count > 0 {
			totalCommits += c.Count
			activeDays++
		}
	}

	activityScore := 0
	if totalCommits > 1000 {
		activityScore += 40
	} else if totalCommits > 500 {
		activityScore += 30
	} else if totalCommits > 200 {
		activityScore += 20
	} else if totalCommits > 50 {
		activityScore += 10
	}

	if activeDays > 200 {
		activityScore += 30
	} else if activeDays > 100 {
		activityScore += 20
	} else if activeDays > 50 {
		activityScore += 10
	}

	// Calculate quality score based on repositories
	var totalStars int
	for _, repo := range repos {
		totalStars += repo.Stars
	}

	qualityScore := 0
	if totalStars > 100 {
		qualityScore += 40
	} else if totalStars > 50 {
		qualityScore += 30
	} else if totalStars > 20 {
		qualityScore += 20
	} else if totalStars > 5 {
		qualityScore += 10
	}

	// Build top repositories (sorted by stars, top 6)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Stars > repos[j].Stars
	})
	topRepos := make(models.JSONArray, 0, 6)
	for i, repo := range repos {
		if i >= 6 {
			break
		}
		topRepos = append(topRepos, map[string]interface{}{
			"name":        repo.Name,
			"full_name":   repo.FullName,
			"description": repo.Description,
			"language":    repo.Language,
			"stars":       repo.Stars,
			"forks":       repo.Forks,
		})
	}

	// Build repo language aggregate
	langCounts := make(map[string]int)
	for _, repo := range repos {
		if repo.Language != "" {
			langCounts[repo.Language]++
		}
	}
	repoLanguages := make(models.JSONMap)
	for lang, count := range langCounts {
		repoLanguages[lang] = count
	}

	// Update profile
	updates := map[string]interface{}{
		"total_commits":    totalCommits,
		"activity_score":   activityScore,
		"quality_score":    qualityScore,
		"overall_score":    (activityScore + qualityScore) / 2,
		"top_repositories": topRepos,
		"repo_languages":   repoLanguages,
	}

	s.githubRepo.UpdateProfileStats(ctx, userID, updates)
}

func (s *GitHubServiceImpl) refreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	tokenSource := s.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	return tokenSource.Token()
}

func (s *GitHubServiceImpl) handlePushWebhook(ctx context.Context, payload map[string]interface{}) error {
	sender, ok := payload["sender"].(map[string]interface{})
	if !ok {
		return nil
	}

	login, ok := sender["login"].(string)
	if !ok {
		return nil
	}

	// Find user by GitHub username
	profile, err := s.githubRepo.GetProfileByUsername(ctx, login)
	if err != nil {
		return nil
	}

	// Trigger sync
	go s.SyncGitHubData(context.Background(), profile.EmployeeID)

	return nil
}

func (s *GitHubServiceImpl) handleStarWebhook(ctx context.Context, payload map[string]interface{}) error {
	fmt.Println("Star webhook received")
	return nil
}

func (s *GitHubServiceImpl) calculateInitialActivityScore(profile *models.GithubProfile) int {
	score := 0

	if profile.PublicRepos >= 20 {
		score += 20
	} else {
		score += profile.PublicRepos
	}

	if profile.Followers >= 100 {
		score += 20
	} else if profile.Followers >= 50 {
		score += 15
	} else if profile.Followers >= 20 {
		score += 10
	} else if profile.Followers >= 5 {
		score += 5
	}

	if profile.AccountAgeDays >= 730 {
		score += 15
	} else if profile.AccountAgeDays >= 365 {
		score += 10
	} else if profile.AccountAgeDays >= 180 {
		score += 5
	}

	if profile.GithubBio != "" {
		if len(profile.GithubBio) > 20 {
			score += 10
		} else {
			score += 5
		}
	}

	if profile.GithubLocation != "" {
		score += 5
	}

	if profile.GithubCompany != "" {
		score += 5
	}

	if profile.IsVerified {
		score += 10
	}

	if score > 100 {
		return 100
	}
	return score
}

func (s *GitHubServiceImpl) calculateInitialQualityScore(profile *models.GithubProfile) int {
	score := 0

	score += 10 // Base for having repos

	if profile.GithubBio != "" {
		if len(profile.GithubBio) > 50 {
			score += 10
		} else {
			score += 5
		}
	}

	if profile.GithubBlog != "" {
		score += 10
	}

	if profile.GithubAvatar != "" {
		score += 10
	}

	if profile.AccountAgeDays >= 730 {
		score += 20
	} else if profile.AccountAgeDays >= 365 {
		score += 15
	} else if profile.AccountAgeDays >= 180 {
		score += 10
	} else if profile.AccountAgeDays >= 90 {
		score += 5
	}

	if profile.IsVerified {
		score += 15
	}

	if score > 100 {
		return 100
	}
	return score
}

// Helper functions
func generateRandomPassword() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
