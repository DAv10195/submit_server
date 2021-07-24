package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/DAv10195/submit_server/db"
	"github.com/DAv10195/submit_server/elements/users"
	"github.com/DAv10195/submit_server/fs"
	"github.com/DAv10195/submit_server/path"
	"github.com/DAv10195/submit_server/server"
	"github.com/DAv10195/submit_server/session"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func newStartCommand(ctx context.Context, args []string) *cobra.Command {
	var setupErr error
	var configFilePath string
	// create the command
	startCmd := &cobra.Command{
		Use: start,
		Short: fmt.Sprintf("%s %s", start, submitServer),
		SilenceUsage: true,
		SilenceErrors: true,
		RunE: func (cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if setupErr != nil {
				return setupErr
			}
			// define logging level and other configuration
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
			// handle the DB dir
			dir := viper.GetString(flagDbDir)
			if err := db.InitDB(dir); err != nil {
				return err
			}
			fsPwd, err := handleConfigEncryption(viper.GetString(flagFileServerPassword), configFilePath)
			if err != nil {
				return err
			}
			// initialize the submit file server client
			if err := fs.Init(viper.GetString(flagFileServerHost), viper.GetInt(flagFileServerPort), viper.GetString(flagFileServerUser), fsPwd, viper.GetBool(flagFsUseTls), viper.GetString(flagTrustedCaFile), viper.GetBool(flagSkipTlsVerify)); err != nil {
				return err
			}
			// initialize the session management
			if err := session.Init(dir); err != nil {
				return err
			}
			// make sure the default admin user exists
			if err := users.InitDefaultAdmin(); err != nil {
				return err
			}
			// run the server
			tlsConf, err := server.GetTlsConfig(viper.GetString(flagTlsCertFile), viper.GetString(flagTlsKeyFile))
			if err != nil {
				return err
			}
			wg := &sync.WaitGroup{}
			wg.Add(1)
			srv := server.InitServer(viper.GetInt(flagServerPort), tlsConf, wg, ctx)
			go func() {
				var serverErr error
				if tlsConf != nil {
					serverErr = srv.ListenAndServeTLS("", "")
				} else {
					serverErr = srv.ListenAndServe()
				}
				if serverErr != http.ErrServerClosed {
					logger.WithError(err).Fatal("submit server crashed")
				}
			}()
			logger.Info("server is running")
			<- ctx.Done()
			logger.Info("stopping server...")
			ctx, timeout := context.WithTimeout(context.Background(), time.Minute)
			defer timeout()
			if err := srv.Shutdown(ctx); err != nil {
				return err
			}
			wg.Wait()
			return nil
		},
	}
	configFlagSet := pflag.NewFlagSet(submit, pflag.ContinueOnError)
	_ = configFlagSet.StringP(flagConfigFile, "c", "", "path to submit server config file")
	configFlagSet.SetOutput(ioutil.Discard)
	_ = configFlagSet.Parse(args[1:])
	configFilePath, _ = configFlagSet.GetString(flagConfigFile)
	if configFilePath == "" {
		configFilePath = filepath.Join(path.GetDefaultConfigDirPath(), defaultConfigFileName)
	}
	viper.SetConfigType(yaml)
	viper.SetConfigFile(configFilePath)
	viper.SetDefault(flagLogFileAndStdout, deLogFileAndStdOut)
	viper.SetDefault(flagLogFileMaxSize, defMaxLogFileSize)
	viper.SetDefault(flagLogFileMaxAge, defMaxLogFileAge)
	viper.SetDefault(flagLogFileMaxBackups, defMaxLogFileBackups)
	viper.SetDefault(flagLogLevel, info)
	viper.SetDefault(flagServerPort, defPort)
	viper.SetDefault(flagDbDir, path.GetDefaultDbDirPath())
	viper.SetDefault(flagFileServerHost, defFileServerHost)
	viper.SetDefault(flagFileServerPort, defFileServerPort)
	viper.SetDefault(flagFileServerUser, defFileServerUser)
	viper.SetDefault(flagFileServerPassword, defFileServerPassword)
	viper.SetDefault(flagSkipTlsVerify, defSkipTlsVerify)
	viper.SetDefault(flagFsUseTls, defFsUseTls)
	startCmd.Flags().AddFlagSet(configFlagSet)
	startCmd.Flags().Int(flagLogFileMaxBackups, viper.GetInt(flagLogFileMaxBackups), "maximum number of log file rotations")
	startCmd.Flags().Int(flagLogFileMaxSize, viper.GetInt(flagLogFileMaxSize), "maximum size of the log file before it's rotated")
	startCmd.Flags().Int(flagLogFileMaxAge, viper.GetInt(flagLogFileMaxAge), "maximum age of the log file before it's rotated")
	startCmd.Flags().Bool(flagLogFileAndStdout, viper.GetBool(flagLogFileAndStdout), "write logs to stdout if log-file is specified?")
	startCmd.Flags().String(flagLogLevel, viper.GetString(flagLogLevel), "logging level [panic, fatal, error, warn, info, debug]")
	startCmd.Flags().String(flagLogFile, viper.GetString(flagLogFile), "log to file, specify the file location")
	startCmd.Flags().String(flagDbDir, viper.GetString(flagDbDir), "db directory of the submit server")
	startCmd.Flags().Int(flagServerPort, viper.GetInt(flagServerPort), "port the submit server should listen on")
	startCmd.Flags().String(flagFileServerHost, viper.GetString(flagFileServerHost), "submit file server hostname (or ip address)")
	startCmd.Flags().Int(flagFileServerPort, viper.GetInt(flagFileServerPort), "submit file server port")
	startCmd.Flags().String(flagFileServerUser, viper.GetString(flagFileServerUser), "user to be used when authenticating against submit file server")
	startCmd.Flags().String(flagFileServerPassword, viper.GetString(flagFileServerPassword), "password to be used when authenticating against submit file server")
	startCmd.Flags().String(flagTlsCertFile, viper.GetString(flagTlsCertFile), "path to a file containing a certificate to use for tls")
	startCmd.Flags().String(flagTlsKeyFile, viper.GetString(flagTlsKeyFile), "path to a file containing a key to use for tls")
	startCmd.Flags().Bool(flagSkipTlsVerify, viper.GetBool(flagSkipTlsVerify), "skip tls verification")
	startCmd.Flags().String(flagTrustedCaFile, viper.GetString(flagTrustedCaFile), "trusted ca bundle path")
	startCmd.Flags().Bool(flagFsUseTls, viper.GetBool(flagFsUseTls), "use tls when accessing submit file server")
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		setupErr = err
	}
	return startCmd
}

// encrypt passwords if they are not encrypted yet
func handleConfigEncryption(pwd, configFilePath string) (string, error) {
	var writeToConfRequired bool
	var err error
	var encryptedPassword string
	if !strings.HasPrefix(pwd, encryptedPrefix) {
		writeToConfRequired = true
		encryptedPassword, err = db.Encrypt(pwd)
		if err != nil {
			return "", err
		}
	} else {
		encryptedPassword = strings.TrimPrefix(pwd, encryptedPrefix)
	}
	if writeToConfRequired {
		if _, err = os.Stat(configFilePath); err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
		} else {
			confLines, err := readConfLines(configFilePath)
			if err != nil {
				return "", err
			}
			for i := 0; i < len(confLines); i++ {
				if strings.Contains(confLines[i], flagFileServerPassword) {
					confLines[i] = fmt.Sprintf("%s: %s%s", flagFileServerPassword, encryptedPrefix, encryptedPassword)
				}
			}
			if err = writeConfLines(confLines, configFilePath); err != nil {
				return "", err
			}
		}
	}
	return encryptedPassword, nil
}

// read the conf lines and return a list of them
func readConfLines(configFilePath string) ([]string, error) {
	confFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := confFile.Close(); err != nil {
			logger.WithError(err).Error("error closing config file after reading")
		}
	}()
	var confLines []string
	confScanner := bufio.NewScanner(confFile)
	for confScanner.Scan() {
		confLines = append(confLines, confScanner.Text())
	}
	return confLines, confScanner.Err()
}

// write the given conf lines to the given path
func writeConfLines(confLines []string, configFilePath string) error {
	confFile, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := confFile.Close(); err != nil {
			logger.WithError(err).Error("error closing config file after writing")
		}
	}()
	confWriter := bufio.NewWriter(confFile)
	for _, line := range confLines {
		if _, err := fmt.Fprintln(confWriter, line); err != nil {
			return err
		}
	}
	return confWriter.Flush()
}
