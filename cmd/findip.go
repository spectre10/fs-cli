package cmd

import (
	"fmt"

	"github.com/spectre10/fileshare-cli/lib"
	"github.com/spf13/cobra"
)

// findipCmd represents the findip command
var findipCmd = &cobra.Command{
	Use:   "findip",
	Short: "find your IP address",
	Long:  `This command finds your IP address using Google STUN servers and the request is made via WebRTC.`,
	Run: func(cmd *cobra.Command, args []string) {
		ip, err := lib.Find()
		if err != nil {
			panic(err)
		}
		if len(ip) == 0 {
			fmt.Println("Could not find the IP address!")
			return
		}
		fmt.Println(ip)
	},
}

func init() {
	rootCmd.AddCommand(findipCmd)
}
