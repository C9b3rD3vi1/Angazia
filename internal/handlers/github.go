package handlers

import (
	"net/http"
	"context"

	"github.com/gofiber/fiber/v2"
	
	"github.com/Angazia/internal/services"
	"github.com/Angazia/internal/pkg/utils"
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
		Secure:   fiber.IsProduction(),
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/",
	})
	
	c.Cookie(&fiber.Cookie{
		Name:     "github_oauth_redirect",
		Value:    redirectTo,
		HTTPOnly: true,
		Secure:   fiber.IsProduction(),
		MaxAge:   600,
		Path:     "/",
	})
	
	authURL := h.githubService.GetAuthURL(state, redirectTo)
	
	if c.Get("X-Requested-With") == "XMLHttpRequest" {
		return c.JSON(fiber.Map{
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_state",
			"message": "State parameter mismatch",
		})
	}
	
	c.ClearCookie("github_oauth_state")
	
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_code",
			"message": "Authorization code not found",
		})
	}
	
	// Exchange code for token
	oauthToken, err := h.githubService.ExchangeCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "token_exchange_failed",
			"message": err.Error(),
		})
	}
	
	// Get current user ID from JWT if logged in
	userID := utils.GetUserIDFromContext(c)
	
	// Handle GitHub login/connection
	result, err := h.githubService.HandleGitHubLogin(c.Context(), oauthToken, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "github_login_failed",
			"message": err.Error(),
		})
	}
	
	redirectURL := c.Cookies("github_oauth_redirect")
	c.ClearCookie("github_oauth_redirect")
	
	if redirectURL == "" {
		redirectURL = "/employee/dashboard"
	}
	
	// Generate JWT token for new users
	var jwtToken string
	if result.IsNewUser {
		jwtToken, err = utils.GenerateJWT(result.UserID, "employee")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "token_generation_failed",
				"message": err.Error(),
			})
		}
		
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    jwtToken,
			HTTPOnly: true,
			Secure:   fiber.IsProduction(),
			MaxAge:   86400,
			Path:     "/",
		})
	}
	
	if c.Get("X-Requested-With") == "XMLHttpRequest" {
		response := fiber.Map{
			"success":      true,
			"redirect_url": redirectURL,
		}
		if result.IsNewUser {
			response["token"] = jwtToken
			response["user"] = fiber.Map{
				"id":    result.UserID,
				"email": result.Email,
			}
		}
		return c.JSON(response)
	}
	
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// Connect connects GitHub to existing account
func (h *GitHubHandler) Connect(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
	}
	
	state := utils.GenerateRandomString(32)
	authURL := h.githubService.GetAuthURL(state, "/employee/dashboard")
	
	return c.JSON(fiber.Map{
		"auth_url": authURL,
		"state":    state,
	})
}

// Disconnect removes GitHub connection
func (h *GitHubHandler) Disconnect(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
	}
	
	if err := h.githubService.DisconnectGitHubAccount(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "disconnect_failed",
			"message": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "GitHub account disconnected successfully",
	})
}

// Sync triggers manual GitHub data sync
func (h *GitHubHandler) Sync(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
	}
	
	// Start async sync
	go h.githubService.SyncGitHubData(c.Context(), userID)
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "GitHub sync started in background",
		"status":  "processing",
	})
}

// GetProfile returns GitHub profile
func (h *GitHubHandler) GetProfile(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
	}
	
	profile, err := h.githubService.GetGitHubProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "GitHub profile not found. Please connect your GitHub account.",
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    profile,
	})
}

// GetRepos returns GitHub repositories
func (h *GitHubHandler) GetRepos(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "fetch_failed",
			"message": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"repos":       repos,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetContributions returns GitHub contribution calendar
func (h *GitHubHandler) GetContributions(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "Please log in first",
		})
	}
	
	days := c.QueryInt("days", 365)
	
	contributions, err := h.githubService.GetGitHubContributions(c.Context(), userID, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "fetch_failed",
			"message": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"data":    contributions,
	})
}

// Webhook handles GitHub webhook events
func (h *GitHubHandler) Webhook(c *fiber.Ctx) error {
	signature := c.Get("X-Hub-Signature-256")
	if signature == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing_signature",
		})
	}
	
	body := c.Body()
	eventType := c.Get("X-GitHub-Event")
	deliveryID := c.Get("X-GitHub-Delivery")
	
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid_payload",
		})
	}
	
	// Process webhook asynchronously
	go h.githubService.ProcessWebhook(context.Background(), eventType, deliveryID, payload)
	
	return c.JSON(fiber.Map{
		"status": "received",
	})
}