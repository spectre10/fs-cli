/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	// "fmt"
	"fmt"

	"github.com/spectre10/fileshare-cli/session/send"
	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "To send a file",
	Long: `This command is used to send a file. For example,
    $ fileshare-cli send --file <PathAndNameOfFile>`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("file")
		if path == "" {
			fmt.Println("Missing file path")
			return
		}
		session := send.NewSession(path)
		err := session.Connect()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.PersistentFlags().String("file", "", "name and path of the file")
}
