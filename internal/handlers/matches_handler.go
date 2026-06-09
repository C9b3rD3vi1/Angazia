package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
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
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	matches, err := h.matchingService.GetJobMatches(c.Context(), userID.(string), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, matches)
}

// GetCandidateMatches returns candidate recommendations for an employer
func (h *MatchingHandler) GetCandidateMatches(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	matches, err := h.matchingService.GetCandidateMatches(c.Context(), jobID, userID.(string), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, matches)
}

// GetDetailedMatchAnalysis returns detailed AI analysis for a specific job-candidate pair
func (h *MatchingHandler) GetDetailedMatchAnalysis(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	employeeID := c.Params("employeeId")
	
	if jobID == "" || employeeID == "" {
		return utils.BadRequest(c, "Both job ID and employee ID are required")
	}
	
	analysis, err := h.matchingService.GetDetailedMatchAnalysis(c.Context(), jobID, employeeID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, analysis)
}

// AnalyzeSkillsGap analyzes skills gap for a candidate
func (h *MatchingHandler) AnalyzeSkillsGap(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}
	
	analysis, err := h.matchingService.AnalyzeSkillsGap(c.Context(), jobID, userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, analysis)
}

// GenerateCoverLetter generates an AI-powered cover letter
func (h *MatchingHandler) GenerateCoverLetter(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req struct {
		JobID string `json:"job_id" validate:"required"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	coverLetter, err := h.matchingService.GenerateCoverLetter(c.Context(), req.JobID, userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, fiber.Map{
		"cover_letter": coverLetter,
	})
}

// GetEmployerMatchAnalysis returns detailed match analysis for an employer viewing a candidate
func (h *MatchingHandler) GetEmployerMatchAnalysis(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	employeeID := c.Params("employeeId")

	if jobID == "" || employeeID == "" {
		return utils.BadRequest(c, "Both job ID and employee ID are required")
	}

	analysis, err := h.matchingService.GetDetailedMatchAnalysis(c.Context(), jobID, employeeID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, analysis)
}

// GetEmployerSkillsGap returns skills gap analysis for an employer viewing a candidate
func (h *MatchingHandler) GetEmployerSkillsGap(c *fiber.Ctx) error {
	jobID := c.Params("jobId")
	employeeID := c.Params("employeeId")

	if jobID == "" || employeeID == "" {
		return utils.BadRequest(c, "Both job ID and employee ID are required")
	}

	analysis, err := h.matchingService.AnalyzeSkillsGap(c.Context(), jobID, employeeID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, analysis)
}

// SubmitMatchFeedback handles employer feedback on a match
func (h *MatchingHandler) SubmitMatchFeedback(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		MatchID  string `json:"match_id" validate:"required"`
		Rating   int    `json:"rating" validate:"required,min=1,max=5"`
		Feedback string `json:"feedback"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if req.MatchID == "" {
		return utils.BadRequest(c, "Match ID is required")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return utils.BadRequest(c, "Rating must be between 1 and 5")
	}

	err := h.matchingService.SubmitMatchFeedback(c.Context(), req.MatchID, userID.(string), req.Feedback, req.Rating)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, fiber.Map{"message": "Feedback submitted"})
}

// GenerateInterviewQuestions generates interview questions for a job
func (h *MatchingHandler) GenerateInterviewQuestions(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}
	
	questions, err := h.matchingService.GenerateInterviewQuestions(c.Context(), jobID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, fiber.Map{
		"questions": questions,
	})
}
