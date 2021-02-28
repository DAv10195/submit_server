package cmd

import (
	"context"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func newStartCommand(ctx context.Context, args []string) *cobra.Command {
	var setupErr error
	startCmd := &cobra.Command{
		Use: "start",
		Short: "start submit server",
		SilenceUsage: true,
		SilenceErrors: true,
		RunE: func (cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if setupErr != nil {
				return setupErr
			}
			logLevel := viper.GetString(flagLogLevel)
			level, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logrus.SetLevel(level)
			logFile := viper.GetString(flagLogFile)
			if logFile != "" {
				lumberjackLogger := &lumberjack.Logger{
					Filename:   viper.GetString(flagLogFile),
					MaxSize:    viper.GetInt(flagLogFileMaxSize),
					MaxBackups: viper.GetInt(flagLogFileMaxBackups),
					MaxAge:     viper.GetInt(flagLogFileMaxAge),
					LocalTime:  true,
				}
				if viper.GetBool(flagLogFileAndStdout) {
					logrus.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
				} else {
					logrus.SetOutput(lumberjackLogger)
				}
			} else {
				logger.Debug("log file undefined")
			}
			if err := db.InitDB(viper.GetString(flagDbDir)); err != nil {
				return err
			}
			if err := users.InitDefaultAdmin(); err != nil {
				return err
			}
			cfg := &server.Config{}
			cfg.Port = viper.GetInt(flagServerPort)
			cfg.NumberOfServerGoroutines = viper.GetInt(flagNumWorkers)
			srv, wg := server.InitServer(ctx, cfg)
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					logger.WithError(err).Error("submit server crashed")
				}
			}()
			logger.Info("server is running")
			<- ctx.Done()
			wg.Wait()
			return nil
		},
	}
	configFlagSet := pflag.NewFlagSet("submit", pflag.ContinueOnError)
	_ = configFlagSet.StringP(flagConfigFile, "c", "", "path to submit server config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(args[1:])
	configFilePath, _ := configFlagSet.GetString(flagConfigFile)
	if configFilePath == "" {
		configFilePath = getDefaultConfigFilePath()
	}
	viper.SetConfigType(yaml)
	viper.SetConfigFile(configFilePath)
	viper.SetDefault(flagLogFileAndStdout, deLogFileAndStdOut)
	viper.SetDefault(flagLogFileMaxSize, defMaxLogFileSize)
	viper.SetDefault(flagLogFileMaxAge, defMaxLogFileAge)
	viper.SetDefault(flagLogFileMaxBackups, defMaxLogFileBackups)
	viper.SetDefault(flagLogLevel, info)
	viper.SetDefault(flagNumWorkers, server.DefNumberOfServerGoroutines)
	viper.SetDefault(flagServerPort, server.DefPort)
	viper.SetDefault(flagDbDir, getDefaultDbDirPath())
	startCmd.Flags().AddFlagSet(configFlagSet)
	startCmd.Flags().Int(flagLogFileMaxBackups, viper.GetInt(flagLogFileMaxBackups), "maximum number of log file rotations")
	startCmd.Flags().Int(flagLogFileMaxSize, viper.GetInt(flagLogFileMaxSize), "maximum size of the log file before it's rotated")
	startCmd.Flags().Int(flagLogFileMaxAge, viper.GetInt(flagLogFileMaxAge), "maximum age of the log file before it's rotated")
	startCmd.Flags().Bool(flagLogFileAndStdout, viper.GetBool(flagLogFileAndStdout), "write logs to stdout if log-file is specified?")
	startCmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	startCmd.Flags().String(flagLogFile, viper.GetString(flagLogFile), "log to file, specify the file location")
	startCmd.Flags().String(flagDbDir, viper.GetString(flagDbDir), "db directory of the submit server")
	startCmd.Flags().Int(flagServerPort, viper.GetInt(flagServerPort), "port the submit server should listen on")
	startCmd.Flags().Int(flagNumWorkers, viper.GetInt(flagNumWorkers), "number of worker goroutines to be started by the submit server")
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		setupErr = err
	}
	return startCmd
}
