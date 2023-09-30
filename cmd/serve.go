package cmd

import (
	"github.com/spectre10/fs-cli/web"
	"github.com/spf13/cobra"
)

// serveCmd represents the server command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start a local server for browser-based UI",
	Long: `start a UI hosted at a particular port.
	for defining port use the "--address" flag
	eg. fs-cli server --address :8080
	by default, it will be port 8080.`,
	Run: func(cmd *cobra.Command, args []string) {
		add, err := cmd.Flags().GetString("address")
		if err != nil {
			panic(err)
		}
		web.StartServer(add)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("address", "", "A help for foo")
}
