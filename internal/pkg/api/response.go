package response

import (
	"errors"
	"log/slog"

	apperrors "hrms/internal/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type Single struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type Paginated struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Pagination PageInfo    `json:"pagination"`
}

type PageInfo struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Option struct {
	Value string      `json:"value"`
	Label string      `json:"label"`
	Extra interface{} `json:"extra,omitempty"`
}

type OptionsData struct {
	Options []Option `json:"options"`
}

func OK(c fiber.Ctx, data interface{}) error {
	return c.JSON(Single{Success: true, Data: data})
}

func OKWithMessage(c fiber.Ctx, data interface{}, message string) error {
	return c.JSON(Single{Success: true, Data: data, Message: message})
}

func Created(c fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Single{Success: true, Data: data})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Paginate(c fiber.Ctx, data interface{}, page, perPage int, total int64) error {
	totalPages := (int(total) + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	return c.JSON(Paginated{
		Success: true,
		Data:    data,
		Pagination: PageInfo{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func Options(c fiber.Ctx, opts []Option) error {
	return c.JSON(Single{
		Success: true,
		Data:    OptionsData{Options: opts},
	})
}

func Error(c fiber.Ctx, err error) error {
	var domainErr *apperrors.DomainError
	if errors.As(err, &domainErr) {
		return c.Status(domainErr.HTTPStatus).JSON(ErrorResponse{
			Success: false,
			Error: ErrorInfo{
				Code:    domainErr.Code,
				Message: domainErr.Message,
			},
		})
	}
	slog.Error("unhandled error in response.Error", "error", err.Error(), "path", c.Path(), "method", c.Method())
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    apperrors.ErrInternal.Code,
			Message: apperrors.ErrInternal.Message,
		},
	})
}
