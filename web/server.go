package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func StartServer(add string) {
	err := initLogger()
	if err != nil {
		log.Fatal(err)
	}
	err = initFS()
	if err != nil {
		logger.Error(fmt.Sprint(err))
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
	err = app.Listen(add)
	logger.Error(fmt.Sprint(err))
	shutdown()
}
