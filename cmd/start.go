package cmd

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use: "start",
	Short: "start submit_server",
	SilenceUsage: true,
	SilenceErrors: true,
	RunE: func (cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
