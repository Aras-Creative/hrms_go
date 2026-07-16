package delivery

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
	storageAdapter "hrms/internal/storage/adapter"
	"hrms/internal/storage/repository"
	"hrms/internal/storage/usecase"
)

type StorageHandler struct {
	uc          *usecase.StorageUsecase
	resolver    repository.URLResolver
	auditLogger *storageAdapter.AuditLogger
}

func New(uc *usecase.StorageUsecase, resolver repository.URLResolver, auditLogger *storageAdapter.AuditLogger) *StorageHandler {
	return &StorageHandler{uc: uc, resolver: resolver, auditLogger: auditLogger}
}

func (h *StorageHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	s := r.Group("/storage")
	s.Post("/upload", authMw, h.Upload)
	s.Post("/replace", authMw, h.Replace)
	s.Get("/:id", authMw, h.Download)
	s.Delete("/:id", authMw, h.Delete)
	s.Get("/:module/:referenceID", authMw, h.ListByModule)
}

func (h *StorageHandler) Upload(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	module := c.FormValue("module")
	if module == "" {
		return response.Error(c, errors.NewInvalidInput("module is required"))
	}

	var referenceID *string
	if ref := c.FormValue("reference_id"); ref != "" {
		referenceID = &ref
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("file is required"))
	}

	fh, err := file.Open()
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to open file"))
	}
	defer fh.Close()

	doc, err := h.uc.Upload(c.Context(), fh, file, userID.(string), module, referenceID)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), userID.(string), doc.ID, "",
			storageAdapter.ActionUpload, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"module": module, "original_name": doc.OriginalName, "size_bytes": doc.SizeBytes},
		)
	}

	return response.Created(c, newDocumentResponse(doc, h.resolver))
}

func (h *StorageHandler) Replace(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	module := c.FormValue("module")
	if module == "" {
		return response.Error(c, errors.NewInvalidInput("module is required"))
	}

	ref := c.FormValue("reference_id")
	if ref == "" {
		return response.Error(c, errors.NewInvalidInput("reference_id is required"))
	}
	referenceID := &ref

	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("file is required"))
	}

	fh, err := file.Open()
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to open file"))
	}
	defer fh.Close()

	doc, err := h.uc.Replace(c.Context(), fh, file, userID.(string), module, referenceID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, newDocumentResponse(doc, h.resolver))
}

func (h *StorageHandler) Download(c fiber.Ctx) error {
	docID := c.Params("id")
	if docID == "" {
		return response.Error(c, errors.NewInvalidInput("document id is required"))
	}

	reader, doc, err := h.uc.Download(c.Context(), docID)
	if err != nil {
		return response.Error(c, err)
	}
	defer reader.Close()

	c.Set("Content-Type", doc.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.OriginalName))

	return c.SendStream(reader, int(doc.SizeBytes))
}

func (h *StorageHandler) Delete(c fiber.Ctx) error {
	docID := c.Params("id")
	if docID == "" {
		return response.Error(c, errors.NewInvalidInput("document id is required"))
	}

	userID := c.Locals("user_id")

	if err := h.uc.Delete(c.Context(), docID); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil && userID != nil {
		h.auditLogger.Log(c.RequestCtx(), userID.(string), docID, "",
			storageAdapter.ActionDelete, c.IP(), string(c.RequestCtx().UserAgent()),
			nil,
		)
	}

	return response.NoContent(c)
}

func (h *StorageHandler) ListByModule(c fiber.Ctx) error {
	module := c.Params("module")
	referenceID := c.Params("referenceID")

	if module == "" || referenceID == "" {
		return response.Error(c, errors.NewInvalidInput("module and reference_id are required"))
	}

	docs, err := h.uc.ListByModule(c.Context(), module, referenceID)
	if err != nil {
		return response.Error(c, err)
	}

	responses := make([]DocumentResponse, len(docs))
	for i := range docs {
		responses[i] = newDocumentResponse(&docs[i], h.resolver)
	}

	return response.OK(c, responses)
}
