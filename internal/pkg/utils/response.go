package utils

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FlashMessage represents a one-time server-rendered banner message
type FlashMessage struct {
	Type    string `json:"type"`    // success, error, warning, info
	Message string `json:"message"`
}

// SetFlash sets a flash message on the given c.Locals for template rendering
func SetFlash(c *fiber.Ctx, msgType, message string) {
	c.Locals("_flash", FlashMessage{Type: msgType, Message: message})
}

// FlashRedirect redirects with flash message as query params (for JS to pick up)
func FlashRedirect(c *fiber.Ctx, urlStr, msgType, message string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return c.Redirect(urlStr)
	}
	q := u.Query()
	q.Set("flash", message)
	q.Set("type", msgType)
	u.RawQuery = q.Encode()
	return c.Redirect(u.String())
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Meta      *MetaData   `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// MetaData contains pagination and additional metadata
type MetaData struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	Count      int   `json:"count,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation errors response
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    MetaData    `json:"meta"`
}

// ResponseBuilder provides a fluent interface for building responses
type ResponseBuilder struct {
	success   bool
	message   string
	data      interface{}
	err       string
	meta      *MetaData
	statusCode int
}

// NewResponse creates a new response builder
func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		success:    true,
		statusCode: http.StatusOK,
	}
}

// SetSuccess sets the success status
func (rb *ResponseBuilder) SetSuccess(success bool) *ResponseBuilder {
	rb.success = success
	return rb
}

// SetMessage sets the response message
func (rb *ResponseBuilder) SetMessage(message string) *ResponseBuilder {
	rb.message = message
	return rb
}

// SetData sets the response data
func (rb *ResponseBuilder) SetData(data interface{}) *ResponseBuilder {
	rb.data = data
	return rb
}

// SetError sets the error message
func (rb *ResponseBuilder) SetError(err string) *ResponseBuilder {
	rb.err = err
	rb.success = false
	return rb
}

// SetMeta sets pagination metadata
func (rb *ResponseBuilder) SetMeta(page, limit int, total int64) *ResponseBuilder {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	rb.meta = &MetaData{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
	return rb
}

// SetStatusCode sets the HTTP status code
func (rb *ResponseBuilder) SetStatusCode(code int) *ResponseBuilder {
	rb.statusCode = code
	return rb
}

// Build builds the final response
func (rb *ResponseBuilder) Build() APIResponse {
	return APIResponse{
		Success:   rb.success,
		Message:   rb.message,
		Data:      rb.data,
		Error:     rb.err,
		Meta:      rb.meta,
		Timestamp: time.Now(),
	}
}

// Send sends the response via Fiber context
func (rb *ResponseBuilder) Send(c *fiber.Ctx) error {
	return c.Status(rb.statusCode).JSON(rb.Build())
}

// ========== SUCCESS RESPONSES ==========

// Success sends a success response with data
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessWithMessage sends a success response with message and optional data
func SuccessWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessCreated sends a 201 Created response
func SuccessCreated(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(http.StatusCreated).JSON(APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessPaginated sends a paginated success response
func SuccessPaginated(c *fiber.Ctx, data interface{}, page, limit int, total int64) error {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return c.Status(http.StatusOK).JSON(PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: MetaData{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// SuccessNoContent sends a 204 No Content response
func SuccessNoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// ========== ERROR RESPONSES ==========

// Error sends an error response
func Error(c *fiber.Ctx, statusCode int, err string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *fiber.Ctx, err string) error {
	return c.Status(http.StatusBadRequest).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *fiber.Ctx, err string) error {
	if err == "" {
		err = "Unauthorized access"
	}
	return c.Status(http.StatusUnauthorized).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *fiber.Ctx, err string) error {
	if err == "" {
		err = "Access forbidden"
	}
	return c.Status(http.StatusForbidden).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// NotFound sends a 404 Not Found error
func NotFound(c *fiber.Ctx, entity string) error {
	err := entity
	if entity == "" {
		err = "Resource not found"
	} else {
		err = entity + " not found"
	}
	return c.Status(http.StatusNotFound).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// Conflict sends a 409 Conflict error
func Conflict(c *fiber.Ctx, err string) error {
	return c.Status(http.StatusConflict).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(c *fiber.Ctx, err string) error {
	if err == "" {
		err = "Internal server error"
	}
	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// ValidationErrors sends validation errors
func ValidationErrors(c *fiber.Ctx, errors []ValidationError) error {
	return c.Status(http.StatusBadRequest).JSON(ValidationErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}

// RateLimitExceeded sends a 429 Too Many Requests error
func RateLimitExceeded(c *fiber.Ctx, retryAfter int) error {
	c.Set("Retry-After", string(rune(retryAfter)))
	return c.Status(http.StatusTooManyRequests).JSON(APIResponse{
		Success:   false,
		Error:     "Rate limit exceeded. Please try again later.",
		Timestamp: time.Now(),
	})
}

// ========== HELPER FUNCTIONS ==========

// ParsePagination parses page and limit from query parameters
func ParsePagination(c *fiber.Ctx) (page, limit int) {
	page = c.QueryInt("page", 1)
	limit = c.QueryInt("limit", 20)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	return page, limit
}

// CalculateTotalPages calculates total pages from total count and limit
func CalculateTotalPages(total int64, limit int) int {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return totalPages
}

// ========== STREAMING RESPONSES ==========

// StreamJSON streams a JSON response (for large datasets)
func StreamJSON(c *fiber.Ctx, data interface{}) error {
	c.Set("Content-Type", "application/json")
	c.Set("Transfer-Encoding", "chunked")
	return c.JSON(data)
}

// ========== FILE RESPONSES ==========

// SendFile sends a file as response
func SendFile(c *fiber.Ctx, filePath, filename string) error {
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.SendFile(filePath)
}

// SendCSV sends CSV data as response
func SendCSV(c *fiber.Ctx, data [][]string, filename string) error {
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename="+filename)
	
	// Convert data to CSV format
	var csvData string
	for _, row := range data {
		for i, cell := range row {
			if i > 0 {
				csvData += ","
			}
			csvData += cell
		}
		csvData += "\n"
	}
	
	return c.SendString(csvData)
}

// ========== WEBHOOK RESPONSES ==========

// WebhookAcknowledgment sends a webhook acknowledgment
func WebhookAcknowledgment(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"status":  "received",
		"message": "Webhook processed successfully",
	})
}