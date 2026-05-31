package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func splitQueryParam(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

type SearchHandler struct {
	searchService services.SearchService
}

func NewSearchHandler(searchService services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchJobs searches for jobs
func (h *SearchHandler) SearchJobs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	filters := h.parseSearchFilters(c)
	
	results, err := h.searchService.SearchJobs(c.Context(), filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "search_failed",
			Message: err.Error(),
		})
	}
	
	// Log search if user is authenticated
	if userID := c.Locals("user_id"); userID != nil {
		go h.searchService.SaveSearchHistory(c.Context(), userID.(string), filters.Keywords, filters, "job", int(results.Total), c.IP(), c.Get("User-Agent"))
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    results,
	})
}

// SearchCandidates searches for candidates (employers only)
func (h *SearchHandler) SearchCandidates(c *fiber.Ctx) error {
	userRole := c.Locals("user_role")
	if userRole != "employer" && userRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(APIResponse{
			Success: false,
			Error:   "forbidden",
			Message: "Only employers can search for candidates",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	filters := h.parseSearchFilters(c)
	
	results, err := h.searchService.SearchCandidates(c.Context(), filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "search_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    results,
	})
}

// SearchCompanies searches for companies
func (h *SearchHandler) SearchCompanies(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	filters := h.parseSearchFilters(c)
	
	results, err := h.searchService.SearchCompanies(c.Context(), filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "search_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    results,
	})
}

// GetJobFacets returns job search facets
func (h *SearchHandler) GetJobFacets(c *fiber.Ctx) error {
	filters := h.parseSearchFilters(c)
	
	facets, err := h.searchService.GetJobFacets(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    facets,
	})
}

// GetSearchHistory returns user's search history
func (h *SearchHandler) GetSearchHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	history, err := h.searchService.GetSearchHistory(c.Context(), userID.(string), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    history,
	})
}

// GetPopularSearches returns popular searches
func (h *SearchHandler) GetPopularSearches(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "7"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	popular, err := h.searchService.GetPopularSearches(c.Context(), days, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    popular,
	})
}

// SaveSearch saves a search
func (h *SearchHandler) SaveSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req struct {
		Name       string                `json:"name"`
		Filters    models.SearchFilters  `json:"filters"`
		EntityType string                `json:"entity_type"`
		Frequency  string                `json:"frequency"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	saved, err := h.searchService.SaveSearch(c.Context(), userID.(string), req.Name, req.Filters, req.EntityType, req.Frequency)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "save_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Search saved successfully",
		Data:    saved,
	})
}

// GetSavedSearches returns user's saved searches
func (h *SearchHandler) GetSavedSearches(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	entityType := c.Query("entity_type", "")
	
	searches, err := h.searchService.GetSavedSearches(c.Context(), userID.(string), entityType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    searches,
	})
}

// DeleteSavedSearch deletes a saved search
func (h *SearchHandler) DeleteSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Search ID is required",
		})
	}
	
	if err := h.searchService.DeleteSavedSearch(c.Context(), id, userID.(string)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Saved search deleted",
	})
}

// RunSavedSearch executes a saved search
func (h *SearchHandler) RunSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Search ID is required",
		})
	}
	
	results, err := h.searchService.RunSavedSearch(c.Context(), id, userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "search_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    results,
	})
}

// AutoComplete returns auto-complete suggestions
func (h *SearchHandler) AutoComplete(c *fiber.Ctx) error {
	prefix := c.Query("q", "")
	if prefix == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_query",
			Message: "Search query is required",
		})
	}
	
	entityType := c.Query("type", "job")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	suggestions, err := h.searchService.AutoComplete(c.Context(), prefix, entityType, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "suggest_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    suggestions,
	})
}

// parseSearchFilters parses query parameters into SearchFilters
func (h *SearchHandler) parseSearchFilters(c *fiber.Ctx) models.SearchFilters {
	var isRemote, isHybrid, isVerified, githubConnected *bool
	
	if remote := c.Query("remote"); remote != "" {
		b := remote == "true"
		isRemote = &b
	}
	if hybrid := c.Query("hybrid"); hybrid != "" {
		b := hybrid == "true"
		isHybrid = &b
	}
	if verified := c.Query("verified"); verified != "" {
		b := verified == "true"
		isVerified = &b
	}
	if github := c.Query("github_connected"); github != "" {
		b := github == "true"
		githubConnected = &b
	}
	
	return models.SearchFilters{
		Keywords:         c.Query("q", ""),
		Location:         c.Query("location", ""),
		Radius:           c.QueryInt("radius", 0),
		IsRemote:         isRemote,
		IsHybrid:         isHybrid,
		JobTitle:         c.Query("job_title", ""),
		CompanyName:      c.Query("company", ""),
		Industry:         c.Query("industry", ""),
		EmploymentType:   c.Query("employment_type", ""),
		ExperienceLevel:  c.Query("experience_level", ""),
		MinSalary:        c.QueryInt("min_salary", 0),
		MaxSalary:        c.QueryInt("max_salary", 0),
		Skills:           splitQueryParam(c.Query("skills")),
		PostedWithin:     c.Query("posted_within", ""),
		MinExperience:    c.QueryInt("min_experience", 0),
		MaxExperience:    c.QueryInt("max_experience", 0),
		EducationLevel:   c.Query("education_level", ""),
		GitHubConnected:  githubConnected,
		MinMatchScore:    c.QueryInt("min_match_score", 0),
		CompanySize:      c.Query("company_size", ""),
		IsVerified:       isVerified,
		SortBy:           c.Query("sort_by", "relevance"),
		SortOrder:        c.Query("sort_order", "desc"),
	}
}