package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"hrms/internal/pkg/config"
)

func New(cfg *config.LogConfig) (*logrus.Logger, error) {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetReportCaller(false)

	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	switch strings.ToLower(cfg.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return log, nil
}

func NewWithWriter(cfg *config.LogConfig, w io.Writer) (*logrus.Logger, error) {
	log, err := New(cfg)
	if err != nil {
		return nil, err
	}
	log.SetOutput(w)
	return log, nil
}
