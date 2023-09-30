package web

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func StartServer(add string) {
	var err error
	logFile, err = os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("error opening file: %v", err)
	}
	defer logFile.Close()
	logger := slog.New(slog.NewTextHandler(logFile, nil))
	err = initFS()
	if err != nil {
		logger.Error("%s", err)
		return
	}

	app := fiber.New()
	app.Use("/send", filesystem.New(filesystem.Config{
		Root:  http.FS(staticSend),
		Index: "send.html",
	}))
	app.Use("/receive", filesystem.New(filesystem.Config{
		Root:  http.FS(staticReceive),
		Index: "receive.html",
	}))
	app.Use("/", filesystem.New(filesystem.Config{
		Root:  http.FS(staticIndex),
		Index: "index.html",
	}))

	if add == "" {
		add = ":8080"
	}
	logger.Error("%s", app.Listen(add))
}
