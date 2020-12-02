package main

import (
	"url_reader/config"
	"url_reader/pkg/app"
)

func main() {
	cfg := config.Get()
	app.Run(cfg)
}
