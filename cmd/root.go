package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:	"submit_server",
	SilenceUsage: true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
	}
}
