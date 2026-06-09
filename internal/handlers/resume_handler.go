package handlers

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type ResumeHandler struct {
	profileService services.ProfileService
}

func NewResumeHandler(profileService services.ProfileService) *ResumeHandler {
	return &ResumeHandler{
		profileService: profileService,
	}
}

// UploadResume handles resume file upload and parsing
func (h *ResumeHandler) UploadResume(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	file, err := c.FormFile("resume")
	if err != nil {
		return utils.BadRequest(c, "Please upload a resume file")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".pdf": true, ".docx": true, ".doc": true}
	if !allowedExts[ext] {
		return utils.BadRequest(c, "Only PDF, DOC, and DOCX files are allowed")
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return utils.BadRequest(c, "File size must be less than 5MB")
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		return utils.InternalServerError(c, "Failed to open file")
	}
	defer f.Close()

	// Parse and update profile
	result, err := h.profileService.ParseAndUpdateProfile(c.Context(), userID.(string), f, file.Filename)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, result.Message, result)
}

// GetProfileCompletion returns profile completion percentage
func (h *ResumeHandler) GetProfileCompletion(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	completion, err := h.profileService.GetProfileCompletion(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, completion)
}

// GetSuggestedSkills returns recommended skills
func (h *ResumeHandler) GetSuggestedSkills(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	skills, err := h.profileService.GetSuggestedSkills(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, fiber.Map{
		"skills": skills,
	})
}

// UploadAvatar handles profile picture upload
func (h *ResumeHandler) UploadAvatar(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return utils.BadRequest(c, "Please upload an avatar image")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return utils.BadRequest(c, "Invalid file type. Allowed: JPG, PNG, WEBP")
	}

	if file.Size > 2*1024*1024 {
		return utils.BadRequest(c, "File size must be less than 2MB")
	}

	f, err := file.Open()
	if err != nil {
		return utils.InternalServerError(c, "Failed to open file")
	}
	defer f.Close()

	avatarURL, err := h.profileService.UploadAvatar(c.Context(), userID.(string), f, file.Filename)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Avatar uploaded successfully", fiber.Map{
		"avatar_url": avatarURL,
	})
}

// GetProfileWizard returns wizard data
func (h *ResumeHandler) GetProfileWizard(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	wizard, err := h.profileService.GetProfileWizard(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, wizard)
}
