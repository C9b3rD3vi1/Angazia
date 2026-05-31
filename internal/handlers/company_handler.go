package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type CompanyHandler struct {
	companyService services.CompanyService
	validator      *validator.Validate
}

func NewCompanyHandler(companyService services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		validator:      validator.New(),
	}
}

// GetCompanyProfile returns the authenticated employer's company profile
func (h *CompanyHandler) GetCompanyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	profile, err := h.companyService.GetCompanyProfile(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    profile,
	})
}

// UpdateCompanyProfile updates the company profile
func (h *CompanyHandler) UpdateCompanyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req services.UpdateCompanyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	profile, err := h.companyService.UpdateCompanyProfile(c.Context(), userID.(string), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "update_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Company profile updated successfully",
		Data:    profile,
	})
}

// UploadCompanyLogo uploads a company logo
func (h *CompanyHandler) UploadCompanyLogo(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	file, err := c.FormFile("logo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_file",
			Message: "Please upload a logo file",
		})
	}
	
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "file_open_failed",
			Message: err.Error(),
		})
	}
	defer f.Close()
	
	logoURL, err := h.companyService.UploadCompanyLogo(c.Context(), userID.(string), f, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "upload_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Logo uploaded successfully",
		Data: fiber.Map{
			"logo_url": logoURL,
		},
	})
}

// SubmitVerification submits company verification documents
func (h *CompanyHandler) SubmitVerification(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req services.VerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	verification, err := h.companyService.SubmitVerification(c.Context(), userID.(string), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "submission_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Verification submitted successfully. Our team will review your documents.",
		Data:    verification,
	})
}

// GetVerificationStatus returns the current verification status
func (h *CompanyHandler) GetVerificationStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	status, err := h.companyService.GetVerificationStatus(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    status,
	})
}

// GetCompanyBadges returns company trust badges
func (h *CompanyHandler) GetCompanyBadges(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		userID := c.Locals("user_id")
		if userID != nil {
			companyID = userID.(string)
		}
	}
	
	badges, err := h.companyService.GetCompanyBadges(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    badges,
	})
}

// SubmitReview submits a company review
func (h *CompanyHandler) SubmitReview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	companyID := c.Params("id")
	if companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_company_id",
			Message: "Company ID is required",
		})
	}
	
	var req services.SubmitReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	review, err := h.companyService.SubmitReview(c.Context(), companyID, userID.(string), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "submission_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Review submitted successfully",
		Data:    review,
	})
}

// GetCompanyReviews returns company reviews
func (h *CompanyHandler) GetCompanyReviews(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_company_id",
			Message: "Company ID is required",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	reviews, err := h.companyService.GetCompanyReviews(c.Context(), companyID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    reviews,
	})
}

// GetReviewStats returns company review statistics
func (h *CompanyHandler) GetReviewStats(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_company_id",
			Message: "Company ID is required",
		})
	}
	
	stats, err := h.companyService.GetReviewStats(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    stats,
	})
}

// InviteTeamMember invites a team member
func (h *CompanyHandler) InviteTeamMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req services.InviteTeamMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	invitation, err := h.companyService.InviteTeamMember(c.Context(), userID.(string), userID.(string), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "invitation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Invitation sent successfully",
		Data:    invitation,
	})
}

// AcceptInvitation accepts a team invitation
func (h *CompanyHandler) AcceptInvitation(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_token",
			Message: "Invitation token is required",
		})
	}
	
	if err := h.companyService.AcceptInvitation(c.Context(), token, userID.(string)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "acceptance_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Invitation accepted successfully",
	})
}

// GetTeamMembers returns company team members
func (h *CompanyHandler) GetTeamMembers(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	members, err := h.companyService.GetTeamMembers(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    members,
	})
}

// RemoveTeamMember removes a team member
func (h *CompanyHandler) RemoveTeamMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	memberID := c.Params("memberId")
	if memberID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_member_id",
			Message: "Member ID is required",
		})
	}
	
	if err := h.companyService.RemoveTeamMember(c.Context(), userID.(string), memberID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "removal_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Team member removed successfully",
	})
}

// GetPublicCompanyProfile returns public company profile
func (h *CompanyHandler) GetPublicCompanyProfile(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_company_id",
			Message: "Company ID is required",
		})
	}
	
	profile, err := h.companyService.GetPublicCompanyProfile(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    profile,
	})
}

// GetCompanyAnalytics returns company analytics
func (h *CompanyHandler) GetCompanyAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	days, _ := strconv.Atoi(c.Query("days", "30"))
	
	analytics, err := h.companyService.GetCompanyAnalytics(c.Context(), userID.(string), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    analytics,
	})
}