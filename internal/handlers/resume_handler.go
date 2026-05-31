package handlers

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	
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
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	file, err := c.FormFile("resume")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_file",
			Message: "Please upload a resume file",
		})
	}
	
	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".pdf": true, ".docx": true, ".doc": true}
	if !allowedExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_file_type",
			Message: "Only PDF, DOC, and DOCX files are allowed",
		})
	}
	
	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "file_too_large",
			Message: "File size must be less than 5MB",
		})
	}
	
	// Open file
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "file_open_failed",
			Message: "Failed to open file",
		})
	}
	defer f.Close()
	
	// Parse and update profile
	result, err := h.profileService.ParseAndUpdateProfile(c.Context(), userID.(string), f, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "parsing_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// GetProfileCompletion returns profile completion percentage
func (h *ResumeHandler) GetProfileCompletion(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	completion, err := h.profileService.GetProfileCompletion(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    completion,
	})
}

// GetSuggestedSkills returns recommended skills
func (h *ResumeHandler) GetSuggestedSkills(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	skills, err := h.profileService.GetSuggestedSkills(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"skills": skills,
		},
	})
}

// GetProfileWizard returns wizard data
func (h *ResumeHandler) GetProfileWizard(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	wizard, err := h.profileService.GetProfileWizard(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    wizard,
	})
}