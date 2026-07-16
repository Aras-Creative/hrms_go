package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/sirupsen/logrus"

	"hrms/internal/pkg/config"
)

func Register(app *fiber.App, log *logrus.Logger, corsCfg *config.CORSConfig) {
	app.Use(NewRecover(log))
	app.Use(NewRequestID())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     split(corsCfg.AllowOrigins),
		AllowHeaders:     split(corsCfg.AllowHeaders),
		AllowMethods:     split(corsCfg.AllowMethods),
		AllowCredentials: true,
	}))
	app.Use(NewLogger(log))
}

func RegisterWithConfig(app *fiber.App, log *logrus.Logger, corsCfg *config.CORSConfig, logCfg LoggerConfig) {
	app.Use(NewRecover(log))
	app.Use(NewRequestID())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     split(corsCfg.AllowOrigins),
		AllowHeaders:     split(corsCfg.AllowHeaders),
		AllowMethods:     split(corsCfg.AllowMethods),
		AllowCredentials: true,
	}))
	app.Use(NewLogger(log, logCfg))
}

func split(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
