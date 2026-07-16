package main

import (
	"flag"

	"hrms/internal/bootstrap"
)

func main() {
	cfgPath := flag.String("config", "config/config.yaml", "path to configuration file")
	flag.Parse()

	bootstrap.Run(*cfgPath)
}
