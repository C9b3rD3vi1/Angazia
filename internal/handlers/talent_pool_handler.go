package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type TalentPoolHandler struct {
	talentPoolService services.TalentPoolService
	validator         *validator.Validate
}

func NewTalentPoolHandler(talentPoolService services.TalentPoolService) *TalentPoolHandler {
	return &TalentPoolHandler{
		talentPoolService: talentPoolService,
		validator:         validator.New(),
	}
}

// CreatePool creates a new talent pool
func (h *TalentPoolHandler) CreatePool(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req models.CreateTalentPoolRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	pool, err := h.talentPoolService.CreatePool(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessCreated(c, "Talent pool created successfully", pool)
}

// GetPool retrieves a talent pool by ID
func (h *TalentPoolHandler) GetPool(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	pool, err := h.talentPoolService.GetPool(c.Context(), poolID, userID.(string))
	if err != nil {
		return utils.NotFound(c, "Talent pool")
	}
	
	return utils.Success(c, pool)
}

// UpdatePool updates a talent pool
func (h *TalentPoolHandler) UpdatePool(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	var req models.UpdateTalentPoolRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	pool, err := h.talentPoolService.UpdatePool(c.Context(), poolID, userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Talent pool updated successfully", pool)
}

// DeletePool deletes a talent pool
func (h *TalentPoolHandler) DeletePool(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	if err := h.talentPoolService.DeletePool(c.Context(), poolID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Talent pool deleted successfully", nil)
}

// ListPools lists all talent pools for the employer
func (h *TalentPoolHandler) ListPools(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	result, err := h.talentPoolService.ListPools(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, result)
}

// AddCandidate adds a candidate to a talent pool
func (h *TalentPoolHandler) AddCandidate(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	var req models.AddCandidateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	candidate, err := h.talentPoolService.AddCandidate(c.Context(), poolID, userID.(string), &req)
	if err != nil {
		if err.Error() == "candidate already exists in this pool" {
			return utils.Conflict(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessCreated(c, "Candidate added to talent pool", candidate)
}

// UpdateCandidate updates a candidate in a talent pool
func (h *TalentPoolHandler) UpdateCandidate(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("poolId")
	candidateID := c.Params("candidateId")
	
	if poolID == "" || candidateID == "" {
		return utils.BadRequest(c, "Pool ID and Candidate ID are required")
	}
	
	var req models.UpdateCandidateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	candidate, err := h.talentPoolService.UpdateCandidate(c.Context(), candidateID, poolID, userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Candidate updated successfully", candidate)
}

// RemoveCandidate removes a candidate from a talent pool
func (h *TalentPoolHandler) RemoveCandidate(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("poolId")
	candidateID := c.Params("candidateId")
	
	if poolID == "" || candidateID == "" {
		return utils.BadRequest(c, "Pool ID and Candidate ID are required")
	}
	
	if err := h.talentPoolService.RemoveCandidate(c.Context(), candidateID, poolID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Candidate removed from talent pool", nil)
}

// ListCandidates lists candidates in a talent pool
func (h *TalentPoolHandler) ListCandidates(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	filters := repository.CandidateFilters{
		Status:       c.Query("status", ""),
		MinMatchScore: c.QueryInt("min_score", 0),
		MaxMatchScore: c.QueryInt("max_score", 100),
	}
	
	if tags := c.Query("tags"); tags != "" {
		// Parse comma-separated tags
		filters.Tags = []string{tags}
	}
	
	result, err := h.talentPoolService.ListCandidates(c.Context(), poolID, userID.(string), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, result)
}

// MarkContacted marks a candidate as contacted
func (h *TalentPoolHandler) MarkContacted(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("poolId")
	candidateID := c.Params("candidateId")
	
	if poolID == "" || candidateID == "" {
		return utils.BadRequest(c, "Pool ID and Candidate ID are required")
	}
	
	if err := h.talentPoolService.MarkCandidateContacted(c.Context(), candidateID, poolID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Candidate marked as contacted", nil)
}

// MarkHired marks a candidate as hired
func (h *TalentPoolHandler) MarkHired(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("poolId")
	candidateID := c.Params("candidateId")
	
	if poolID == "" || candidateID == "" {
		return utils.BadRequest(c, "Pool ID and Candidate ID are required")
	}
	
	if err := h.talentPoolService.MarkCandidateHired(c.Context(), candidateID, poolID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Candidate marked as hired", nil)
}

// GetPoolStats gets statistics for a talent pool
func (h *TalentPoolHandler) GetPoolStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	stats, err := h.talentPoolService.GetPoolStats(c.Context(), poolID, userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, stats)
}

// GetEmployerStats gets overall talent pool statistics for the employer
func (h *TalentPoolHandler) GetEmployerStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	stats, err := h.talentPoolService.GetEmployerStats(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, stats)
}

// SearchCandidates searches for candidates in a talent pool
func (h *TalentPoolHandler) SearchCandidates(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	poolID := c.Params("id")
	if poolID == "" {
		return utils.BadRequest(c, "Pool ID is required")
	}
	
	query := c.Query("q", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	if query == "" {
		return utils.BadRequest(c, "Search query is required")
	}
	
	result, err := h.talentPoolService.SearchCandidates(c.Context(), poolID, userID.(string), query, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, result)
}
