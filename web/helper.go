package web

import (
	"embed"
	"io/fs"
	"os"
)

var (
	//go:embed static/*
	static        embed.FS
	staticIndex   fs.FS
	staticSend    fs.FS
	staticReceive fs.FS
	logFile       *os.File
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
