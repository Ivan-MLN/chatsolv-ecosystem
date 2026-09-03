package knowledge

import (
	"authbackend/internal/access"
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service} }

type textRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
}
type faqItemRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
type faqRequest struct {
	WorkspaceID string           `json:"workspace_id"`
	Title       string           `json:"title"`
	FAQs        []faqItemRequest `json:"faqs"`
}
type updateRequest struct {
	Title string `json:"title"`
}

type sourceResponse struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Type             string     `json:"type"`
	Status           string     `json:"status"`
	OriginalFilename string     `json:"original_filename,omitempty"`
	FileSize         int64      `json:"file_size,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type sourceDetailResponse struct {
	sourceResponse
	Content string `json:"content,omitempty"`
}

func publicSource(record SourceRecord) sourceResponse {
	result := sourceResponse{
		ID: record.ID, Title: record.Title, Type: record.Type, Status: record.Status,
		OriginalFilename: record.OriginalFilename, FileSize: record.FileSize,
	}
	if !record.CreatedAt.IsZero() {
		createdAt := record.CreatedAt.UTC()
		result.CreatedAt = &createdAt
	}
	if !record.UpdatedAt.IsZero() {
		updatedAt := record.UpdatedAt.UTC()
		result.UpdatedAt = &updatedAt
	}
	return result
}

func publicSourceDetail(record SourceRecord) sourceDetailResponse {
	detail := sourceDetailResponse{sourceResponse: publicSource(record)}
	if strings.HasPrefix(record.ObjectKey, "inline:") {
		detail.Content = strings.TrimPrefix(record.ObjectKey, "inline:")
	}
	return detail
}

func publicSources(records []SourceRecord) []sourceResponse {
	result := make([]sourceResponse, 0, len(records))
	for _, record := range records {
		result = append(result, publicSource(record))
	}
	return result
}

func (h *Handler) UploadDocument(c *fiber.Ctx) error {
	workspaceID := c.FormValue("workspace_id")
	title := c.FormValue("title")
	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, 400, "Document file is required", "VALIDATION_ERROR")
	}
	opened, err := file.Open()
	if err != nil {
		return response.Fail(c, 400, "Cannot read document", "INVALID_DOCUMENT")
	}
	defer opened.Close()
	mimeType := detectUploadMIME(file, opened)
	_, _ = opened.Seek(0, 0)
	resolved, _ := access.FromLocals(c)
	record, err := h.service.UploadDocument(c.UserContext(), Upload{WorkspaceID: workspaceID, UserID: auth.AuthenticatedUserID(c), Title: title, Filename: file.Filename, MIMEType: mimeType, Size: file.Size, Reader: opened, Unlimited: resolved.Entitlements.IsUnlimited})
	if err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, 202, "Knowledge ingestion queued", publicSource(record))
}
func (h *Handler) CreateText(c *fiber.Ctx) error {
	var in textRequest
	if err := c.BodyParser(&in); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	resolved, _ := access.FromLocals(c)
	record, err := h.service.CreateText(c.UserContext(), auth.AuthenticatedUserID(c), in.WorkspaceID, in.Title, in.Content, "text", resolved.Entitlements.IsUnlimited)
	if err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, 202, "Knowledge ingestion queued", publicSource(record))
}
func (h *Handler) CreateFAQ(c *fiber.Ctx) error {
	var in faqRequest
	if err := c.BodyParser(&in); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	if len(in.FAQs) == 0 {
		return response.Fail(c, 400, "At least one FAQ is required", "VALIDATION_ERROR")
	}
	var body strings.Builder
	for index, faq := range in.FAQs {
		question := strings.TrimSpace(faq.Question)
		answer := strings.TrimSpace(faq.Answer)
		if question == "" || answer == "" {
			return response.Fail(c, 400, "FAQ question and answer are required", "VALIDATION_ERROR")
		}
		_, _ = fmt.Fprintf(&body, "## Question %d\n%s\n\n## Answer %d\n%s\n\n", index+1, question, index+1, answer)
	}
	resolved, _ := access.FromLocals(c)
	record, err := h.service.CreateText(c.UserContext(), auth.AuthenticatedUserID(c), in.WorkspaceID, in.Title, body.String(), "faq", resolved.Entitlements.IsUnlimited)
	if err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, 202, "FAQ ingestion queued", publicSource(record))
}
func (h *Handler) List(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if _, err := h.service.repository.ResolveWorkspace(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID); err != nil {
		return knowledgeError(c, err)
	}
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	items, err := h.service.repository.List(c.UserContext(), workspaceID, limit, 0)
	if err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, 200, "Knowledge retrieved", publicSources(items))
}
func (h *Handler) Get(c *fiber.Ctx) error {
	item, err := h.service.repository.GetForMember(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"))
	if err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, 200, "Knowledge retrieved", publicSourceDetail(item))
}
func (h *Handler) Update(c *fiber.Ctx) error {
	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || mediaType != fiber.MIMEApplicationJSON {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var request updateRequest
	if err = c.BodyParser(&request); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	if err = h.service.Update(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"), request.Title); err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, fiber.StatusAccepted, "Knowledge re-ingestion queued", fiber.Map{"status": "queued"})
}
func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id")); err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, fiber.StatusAccepted, "Knowledge deletion queued", fiber.Map{"status": "deleting"})
}
func (h *Handler) Retry(c *fiber.Ctx) error {
	if err := h.service.Retry(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id")); err != nil {
		return knowledgeError(c, err)
	}
	return response.OK(c, fiber.StatusAccepted, "Knowledge ingestion retry queued", fiber.Map{"status": "queued"})
}
func detectUploadMIME(file *multipart.FileHeader, opened multipart.File) string {
	buffer := make([]byte, 512)
	n, _ := opened.Read(buffer)
	value := file.Header.Get("Content-Type")
	if value == "" || value == "application/octet-stream" {
		value = detectMIME(buffer[:n], file.Filename)
	}
	return value
}
func detectMIME(data []byte, filename string) string {
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "application/pdf"
	}
	if len(data) >= 2 && string(data[:2]) == "PK" {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	if len(data) > 0 {
		if len(filename) >= 4 && filename[len(filename)-4:] == ".csv" {
			return "text/csv"
		}
		return "text/plain"
	}
	return "application/octet-stream"
}
func knowledgeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrForbidden):
		return response.Fail(c, 403, "Knowledge access forbidden", "FORBIDDEN")
	case errors.Is(err, ErrDocumentTooLarge):
		return response.Fail(c, 400, "Document is too large", "DOCUMENT_TOO_LARGE")
	case errors.Is(err, ErrDocumentLimit):
		return response.Fail(c, 409, "Document limit reached", "DOCUMENT_LIMIT_REACHED")
	case errors.Is(err, ErrStorageLimit):
		return response.Fail(c, 409, "Knowledge storage limit reached", "STORAGE_LIMIT_REACHED")
	case errors.Is(err, ErrUnsupportedType):
		return response.Fail(c, 400, "Document type is unsupported", "DOCUMENT_TYPE_UNSUPPORTED")
	case errors.Is(err, ErrDocumentDuplicate):
		return response.Fail(c, 409, "Document already exists", "DOCUMENT_ALREADY_EXISTS")
	case errors.Is(err, ErrInvalidInput):
		return response.Fail(c, 400, "Invalid knowledge input", "VALIDATION_ERROR")
	default:
		return response.Fail(c, 500, "Knowledge operation failed", "INTERNAL_ERROR")
	}
}
