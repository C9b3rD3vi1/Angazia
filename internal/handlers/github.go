package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

type GitHubHandler struct {
	githubService services.GitHubService
}

func NewGitHubHandler(githubService services.GitHubService) *GitHubHandler {
	return &GitHubHandler{
		githubService: githubService,
	}
}

// Login initiates GitHub OAuth flow
func (h *GitHubHandler) Login(c *fiber.Ctx) error {
	state := utils.GenerateRandomString(32)
	redirectTo := c.Query("redirect_to", "/employee/dashboard")

	// Store state in cookie
	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   utils.IsProduction(),
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_redirect",
		Value:    redirectTo,
		HTTPOnly: true,
		Secure:   utils.IsProduction(),
		MaxAge:   600,
		Path:     "/",
	})

	authURL := h.githubService.GetAuthURL(state, redirectTo)

	if c.Get("X-Requested-With") == "XMLHttpRequest" {
		return utils.Success(c, fiber.Map{
			"auth_url": authURL,
			"state":    state,
		})
	}

	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// Callback handles GitHub OAuth callback
func (h *GitHubHandler) Callback(c *fiber.Ctx) error {
	state := c.Query("state")
	cookieState := c.Cookies("github_oauth_state")

	if state == "" || cookieState == "" || state != cookieState {
		return utils.BadRequest(c, "State parameter mismatch")
	}

	c.ClearCookie("github_oauth_state")

	code := c.Query("code")
	if code == "" {
		return utils.BadRequest(c, "Authorization code not found")
	}

	// Exchange code for token
	oauthToken, err := h.githubService.ExchangeCode(c.Context(), code)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	// Get current user ID from JWT if logged in
	userID := utils.GetUserIDFromContext(c)

	// Handle GitHub login/connection
	result, err := h.githubService.HandleGitHubLogin(c.Context(), oauthToken, userID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	redirectURL := c.Cookies("github_oauth_redirect")
	c.ClearCookie("github_oauth_redirect")

	if redirectURL == "" {
		redirectURL = "/employee/dashboard"
	}

	// Generate JWT token for new users
	var jwtToken string
	if result.IsNewUser {
		jwtToken, err = utils.GenerateJWT(result.UserID, "employee", result.Email)
		if err != nil {
			return utils.InternalServerError(c, err.Error())
		}

		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    jwtToken,
			HTTPOnly: true,
			Secure:   utils.IsProduction(),
			MaxAge:   86400,
			Path:     "/",
		})
	}

	if c.Get("X-Requested-With") == "XMLHttpRequest" {
		responseData := fiber.Map{
			"redirect_url": redirectURL,
		}
		if result.IsNewUser {
			responseData["token"] = jwtToken
			responseData["user"] = fiber.Map{
				"id":    result.UserID,
				"email": result.Email,
			}
		}
		return utils.Success(c, responseData)
	}

	return c.Redirect(redirectURL, fiber.StatusFound)
}

// Connect connects GitHub to existing account
func (h *GitHubHandler) Connect(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	state := utils.GenerateRandomString(32)
	redirectTo := c.Query("redirect_to", "/employee/skills")

	// Store state in cookie (required for callback validation)
	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   utils.IsProduction(),
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_redirect",
		Value:    redirectTo,
		HTTPOnly: true,
		Secure:   utils.IsProduction(),
		MaxAge:   600,
		Path:     "/",
	})

	authURL := h.githubService.GetAuthURL(state, redirectTo)

	return utils.Success(c, fiber.Map{
		"auth_url": authURL,
		"state":    state,
	})
}

// Disconnect removes GitHub connection
func (h *GitHubHandler) Disconnect(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	if err := h.githubService.DisconnectGitHubAccount(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "GitHub account disconnected successfully", nil)
}

// Sync triggers manual GitHub data sync
func (h *GitHubHandler) Sync(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	// Start async sync with background context (request context cancelled after response)
	go h.githubService.SyncGitHubData(context.Background(), userID)

	return utils.SuccessWithMessage(c, "GitHub sync started in background", fiber.Map{
		"status": "processing",
	})
}

// GetProfile returns GitHub profile
func (h *GitHubHandler) GetProfile(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	profile, err := h.githubService.GetGitHubProfile(c.Context(), userID)
	if err != nil {
		return utils.NotFound(c, "GitHub profile")
	}

	return utils.Success(c, profile)
}

// GetRepos returns GitHub repositories
func (h *GitHubHandler) GetRepos(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	// Parse filters
	filters := make(map[string]interface{})
	if language := c.Query("language"); language != "" {
		filters["language"] = language
	}
	if fork := c.Query("fork"); fork != "" {
		filters["is_fork"] = fork == "true"
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if sortBy := c.Query("sort"); sortBy != "" {
		filters["sort_by"] = sortBy
		filters["sort_order"] = c.Query("order", "desc")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}

	repos, total, err := h.githubService.GetGitHubRepos(c.Context(), userID, filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, fiber.Map{
		"repos":       repos,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

// GetContributions returns GitHub contribution calendar
func (h *GitHubHandler) GetContributions(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return utils.Unauthorized(c, "Please log in first")
	}

	days := c.QueryInt("days", 365)

	contributions, err := h.githubService.GetGitHubContributions(c.Context(), userID, days)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, contributions)
}

// Webhook handles GitHub webhook events
func (h *GitHubHandler) Webhook(c *fiber.Ctx) error {
	signature := c.Get("X-Hub-Signature-256")
	if signature == "" {
		return utils.Unauthorized(c, "Missing signature")
	}

	body := c.Body()
	eventType := c.Get("X-GitHub-Event")
	deliveryID := c.Get("X-GitHub-Delivery")

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return utils.BadRequest(c, "Invalid payload")
	}

	// Process webhook asynchronously
	go h.githubService.ProcessWebhook(context.Background(), eventType, deliveryID, payload)

	return utils.Success(c, fiber.Map{
		"status": "received",
	})
}
