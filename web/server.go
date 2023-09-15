package web

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

var (
	//go:embed static/index/*
	staticIndex embed.FS
)

func StartServer(add string) error {
	app := fiber.New()
	app.Use("/", filesystem.New(filesystem.Config{
		Root:  http.FS(staticIndex),
		Index: "static/index/index.html",
	}))
	if add == "" {
		add = ":8080"
	}
	return app.Listen(add)
}
