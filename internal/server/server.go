package server

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"hrms/internal/pkg/config"
	"hrms/internal/pkg/middleware"
)

// structValidator wraps go-playground/validator to satisfy fiber.StructValidator.
type structValidator struct {
	v *validator.Validate
}

func (sv *structValidator) Validate(out any) error {
	return sv.v.Struct(out)
}

func New(cfg *config.ServerConfig, log *logrus.Logger) *fiber.App {
	readTimeout, _ := time.ParseDuration(cfg.ReadTimeout)

	app := fiber.New(fiber.Config{
		ReadTimeout:     readTimeout,
		WriteTimeout:    0,
		ErrorHandler:    middleware.NewErrorHandler(log),
		StructValidator: &structValidator{v: validator.New()},
	})

	middleware.Register(app, log, &cfg.CORS)

	return app
}

func Listen(app *fiber.App, cfg *config.ServerConfig, log *logrus.Logger) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("server starting on %s", addr)
	return app.Listen(addr)
}
