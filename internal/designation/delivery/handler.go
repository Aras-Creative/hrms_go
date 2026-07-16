package delivery

import (
	"context"

	"github.com/gofiber/fiber/v3"

	desigAdapter "hrms/internal/designation/adapter"
	"hrms/internal/designation/entity"
	"hrms/internal/designation/models"
	"hrms/internal/designation/usecase"
	response "hrms/internal/pkg/api"
)

type PhotoURLResolver interface {
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

type DesignationHandler struct {
	uc            *usecase.DesignationUsecase
	photoResolver PhotoURLResolver
	auditLogger   *desigAdapter.AuditLogger
}

func NewDesignationHandler(uc *usecase.DesignationUsecase, photoResolver PhotoURLResolver, auditLogger *desigAdapter.AuditLogger) *DesignationHandler {
	return &DesignationHandler{uc: uc, photoResolver: photoResolver, auditLogger: auditLogger}
}

func (h *DesignationHandler) Create(c fiber.Ctx) error {
	var req CreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	d, err := h.uc.Create(c.RequestCtx(), req.Name)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, d.ID, "", desigAdapter.ActionCreate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name},
			)
		}
	}
	return response.Created(c, toResponse(d))
}

func (h *DesignationHandler) FindByID(c fiber.Ctx) error {
	id := c.Params("id")
	d, err := h.uc.FindByID(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toResponse(d))
}

func (h *DesignationHandler) FindAll(c fiber.Ctx) error {
	list, err := h.uc.FindAll(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]DesignationResponse, 0, len(list))
	for _, d := range list {
		resp = append(resp, toResponseFromReadModel(d))
	}

	// Resolve profile photo URLs
	var photoIDs []string
	for _, r := range resp {
		for _, m := range r.Members {
			if m.ProfilePhotoID != nil && *m.ProfilePhotoID != "" {
				photoIDs = append(photoIDs, *m.ProfilePhotoID)
			}
		}
	}
	if len(photoIDs) > 0 && h.photoResolver != nil {
		urls, err := h.photoResolver.ResolveURLs(c.RequestCtx(), photoIDs)
		if err == nil {
			for i := range resp {
				for j := range resp[i].Members {
					if resp[i].Members[j].ProfilePhotoID != nil {
						if url, ok := urls[*resp[i].Members[j].ProfilePhotoID]; ok {
							resp[i].Members[j].ProfilePhotoURL = url
						}
					}
				}
			}
		}
	}

	return response.OK(c, resp)
}

func (h *DesignationHandler) Options(c fiber.Ctx) error {
	opts := h.uc.Options(c.RequestCtx())
	return response.Options(c, opts)
}

func (h *DesignationHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	d, err := h.uc.Update(c.RequestCtx(), id, req.Name)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, id, "", desigAdapter.ActionUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name},
			)
		}
	}
	return response.OK(c, toResponse(d))
}

func (h *DesignationHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.Delete(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, id, "", desigAdapter.ActionDelete,
				c.IP(), string(c.RequestCtx().UserAgent()),
				nil,
			)
		}
	}
	return response.NoContent(c)
}

func toResponse(d *entity.Designation) DesignationResponse {
	return DesignationResponse{
		ID:        d.ID,
		Code:      d.Code,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toResponseFromReadModel(d models.DesignationReadModel) DesignationResponse {
	members := make([]MemberItem, len(d.Members))
	for i, m := range d.Members {
		members[i] = MemberItem{
			ID:             m.ID,
			FullName:       m.FullName,
			EmployeeNumber: m.EmployeeNumber,
			ProfilePhotoID: m.ProfilePhotoID,
		}
	}
	return DesignationResponse{
		ID:          d.ID,
		Code:        d.Code,
		Name:        d.Name,
		Members:     members,
		MemberCount: d.MemberCount,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
