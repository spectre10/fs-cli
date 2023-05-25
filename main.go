package main

import (
	"github.com/spectre10/fileshare-cli/session"
)

func main() {
	sess := session.NewSession()
	sess.Connect()
}
