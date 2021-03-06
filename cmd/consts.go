package cmd

const (
	submit					= "submit"
	submitServer 			= "submit server"
	start					= "start"

	defaultConfigFileName	= "submit_server.yml"
	yaml					= "yaml"
	info					= "info"
	defMaxLogFileSize		= 10
	defMaxLogFileAge		= 3
	defMaxLogFileBackups	= 3
	deLogFileAndStdOut		= false

	flagConfigFile        	= "config-file"
	flagDbDir             	= "db-dir"
	flagServerPort        	= "server-port"
	flagLogLevel          	= "log-level"
	flagLogFile           	= "log-file"
	flagLogFileAndStdout  	= "log-file-and-stdout"
	flagLogFileMaxSize    	= "log-file-max-size"
	flagLogFileMaxBackups 	= "log-file-max-backups"
	flagLogFileMaxAge     	= "log-file-max-age"
)