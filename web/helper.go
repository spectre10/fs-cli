package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

var (
	//go:embed static/*
	static        embed.FS
	staticIndex   fs.FS
	staticSend    fs.FS
	staticReceive fs.FS
	logFile       *os.File
	logger        *slog.Logger
)

func initFS() error {
	var err error
	staticIndex, err = fs.Sub(static, "static/index")
	if err != nil {
		return err
	}

	staticSend, err = fs.Sub(static, "static/send")
	if err != nil {
		return err
	}

	staticReceive, err = fs.Sub(static, "static/receive")
	if err != nil {
		return err
	}
	return nil
}

func initLogger() error {
	var err error
	logFile, err = os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logger = slog.New(slog.NewTextHandler(logFile, nil))
	return nil
}

func shutdown() {
	err := logFile.Close()
	if err != nil {
		logger.Error(fmt.Sprint(err))
	}
}
