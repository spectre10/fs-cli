package cmd

import (
	"fmt"
	"os"

	"github.com/spectre10/fileshare-cli/session/receive"
	"github.com/spf13/cobra"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "To receive a file",
	Long: `Receive a file via this command. 
    For example,$ fileshare-cli receive --file <PathAndNameOfFile>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Paste the SDP:")
		name, _ := cmd.Flags().GetString("file")
		if name == "" {
			fmt.Println("Enter a filename")
			return
		}
		file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		sess := receive.NewSession(file)
		err = sess.Connect()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)
	receiveCmd.PersistentFlags().String("file", "", "name and path of the file")
}
