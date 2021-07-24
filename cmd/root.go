package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func NewRootCmd(ctx context.Context, args []string) *cobra.Command {
	// create a root submit server CLI command and register sub commands
	rootCmd := &cobra.Command{
		Use: submitServer,
		Short: submitServer,
		SilenceUsage: true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newStartCommand(ctx, args))
	// register to env variables
	viper.SetEnvPrefix(submit)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	return rootCmd
}
