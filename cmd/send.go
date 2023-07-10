package cmd

import (
	"fmt"

	"github.com/spectre10/fileshare-cli/session/send"
	"github.com/spf13/cobra"
)

var arr []string

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "To send a file",
	Long: `This command is used to send a file. For example,
    $ fileshare-cli send --file <PathAndNameOfFile>`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(args[0])
		// path, _ := cmd.Flags().GetString("file")
		for i := 0; i < len(arr); i++ {
			fmt.Println(arr[i])
		}
		if len(args) == 0 {
			fmt.Println("Missing file path")
			return
		}
		session := send.NewSession(len(args))
		err := session.Connect(args)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	// sendCmd.PersistentFlags().StringP("file", "f", "", "name and path of the file")
	// sendCmd.PersistentFlags().StringSliceVarP(&arr, "hel", "x", []string{}, "path of file")
}
