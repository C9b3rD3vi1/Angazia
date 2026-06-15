package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
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

func (h *CompanyHandler) GetCompanyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	profile, err := h.companyService.GetCompanyProfile(c.Context(), userID.(string))
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, profile)
}

func (h *CompanyHandler) UpdateCompanyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.UpdateCompanyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	profile, err := h.companyService.UpdateCompanyProfile(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Company profile updated successfully", profile)
}

func (h *CompanyHandler) UploadCompanyLogo(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	file, err := c.FormFile("logo")
	if err != nil {
		return utils.BadRequest(c, "Please upload a logo file")
	}

	f, err := file.Open()
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	defer f.Close()

	logoURL, err := h.companyService.UploadCompanyLogo(c.Context(), userID.(string), f, file.Filename)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Logo uploaded successfully", fiber.Map{
		"logo_url": logoURL,
	})
}

func (h *CompanyHandler) SubmitVerification(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.VerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	verification, err := h.companyService.SubmitVerification(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Verification submitted successfully. Our team will review your documents.", verification)
}

func (h *CompanyHandler) GetVerificationStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	status, err := h.companyService.GetVerificationStatus(c.Context(), userID.(string))
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, status)
}

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
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, badges)
}

func (h *CompanyHandler) SubmitReview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	var req services.SubmitReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	review, err := h.companyService.SubmitReview(c.Context(), companyID, userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Review submitted successfully", review)
}

func (h *CompanyHandler) GetCompanyReviews(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	result, err := h.companyService.GetCompanyReviews(c.Context(), companyID, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}

func (h *CompanyHandler) GetReviewStats(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	stats, err := h.companyService.GetReviewStats(c.Context(), companyID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, stats)
}

func (h *CompanyHandler) InviteTeamMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.InviteTeamMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	invitation, err := h.companyService.InviteTeamMember(c.Context(), userID.(string), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Invitation sent successfully", invitation)
}

func (h *CompanyHandler) AcceptInvitation(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	token := c.Params("token")
	if token == "" {
		return utils.BadRequest(c, "Invitation token is required")
	}

	if err := h.companyService.AcceptInvitation(c.Context(), token, userID.(string)); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Invitation accepted successfully", nil)
}

func (h *CompanyHandler) GetTeamMembers(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	members, err := h.companyService.GetTeamMembers(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, members)
}

func (h *CompanyHandler) RemoveTeamMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return utils.BadRequest(c, "Member ID is required")
	}

	if err := h.companyService.RemoveTeamMember(c.Context(), userID.(string), memberID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Team member removed successfully", nil)
}

func (h *CompanyHandler) UpdateTeamMemberRole(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return utils.BadRequest(c, "Member ID is required")
	}

	var req struct {
		Role string `json:"role" validate:"required,oneof=admin recruiter viewer"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.companyService.UpdateTeamMemberRole(c.Context(), userID.(string), memberID, req.Role); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Team member role updated successfully", nil)
}

func (h *CompanyHandler) ListPendingInvitations(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	invitations, err := h.companyService.ListPendingInvitations(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, invitations)
}

func (h *CompanyHandler) CancelInvitation(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	invitationID := c.Params("invitationId")
	if invitationID == "" {
		return utils.BadRequest(c, "Invitation ID is required")
	}

	if err := h.companyService.CancelInvitation(c.Context(), invitationID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Invitation cancelled successfully", nil)
}

func (h *CompanyHandler) GetPublicCompanyProfile(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	profile, err := h.companyService.GetPublicCompanyProfile(c.Context(), companyID)
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, profile)
}

func (h *CompanyHandler) GetCompanyAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))

	analytics, err := h.companyService.GetCompanyAnalytics(c.Context(), userID.(string), days)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, analytics)
}
