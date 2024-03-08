package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

// Represent the root command without any args.
var rootCmd = &cobra.Command{
	Use:     "fs-cli",
	Short:   "Peer-to-Peer filesharing CLI application",
	Long:    `A Peer-to-Peer multi-threaded filesharing CLI app using WebRTC.`,
	Version: "v0.5.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for s := range sig {
			fmt.Println(s.String())
			os.Exit(0)
		}
	}()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
