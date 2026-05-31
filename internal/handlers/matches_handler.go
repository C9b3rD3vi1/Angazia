package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type MatchingHandler struct {
	matchingService services.MatchingService
}

func NewMatchingHandler(matchingService services.MatchingService) *MatchingHandler {
	return &MatchingHandler{
		matchingService: matchingService,
	}
}

// GetJobMatches returns job recommendations for a candidate
func (h *MatchingHandler) GetJobMatches(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	matches, err := h.matchingService.GetJobMatches(c.Context(), userID.(string), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "matching_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    matches,
	})
}

// GetCandidateMatches returns candidate recommendations for an employer
func (h *MatchingHandler) GetCandidateMatches(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_job_id",
			Message: "Job ID is required",
		})
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	matches, err := h.matchingService.GetCandidateMatches(c.Context(), jobID, userID.(string), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "matching_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    matches,
	})
}

// GetDetailedMatchAnalysis returns detailed AI analysis for a specific job-candidate pair
func (h *MatchingHandler) GetDetailedMatchAnalysis(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	employeeID := c.Params("employeeId")
	
	if jobID == "" || employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_ids",
			Message: "Both job ID and employee ID are required",
		})
	}
	
	analysis, err := h.matchingService.GetDetailedMatchAnalysis(c.Context(), jobID, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "analysis_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    analysis,
	})
}

// AnalyzeSkillsGap analyzes skills gap for a candidate
func (h *MatchingHandler) AnalyzeSkillsGap(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_job_id",
			Message: "Job ID is required",
		})
	}
	
	analysis, err := h.matchingService.AnalyzeSkillsGap(c.Context(), jobID, userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "analysis_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    analysis,
	})
}

// GenerateCoverLetter generates an AI-powered cover letter
func (h *MatchingHandler) GenerateCoverLetter(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req struct {
		JobID string `json:"job_id" validate:"required"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	coverLetter, err := h.matchingService.GenerateCoverLetter(c.Context(), req.JobID, userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "generation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"cover_letter": coverLetter,
		},
	})
}

// GenerateInterviewQuestions generates interview questions for a job
func (h *MatchingHandler) GenerateInterviewQuestions(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_job_id",
			Message: "Job ID is required",
		})
	}
	
	questions, err := h.matchingService.GenerateInterviewQuestions(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "generation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"questions": questions,
		},
	})
}