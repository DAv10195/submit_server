package cmd

const (
	submit					= "submit"
	submitServer 			= "submit_server"
	start					= "start"

	defaultConfigFileName	= "submit_server.yml"
	yaml					= "yaml"
	encryptedPrefix			= "encrypted:"
	info					= "info"
	defPort					= 8080
	defMaxLogFileSize		= 10
	defMaxLogFileAge		= 3
	defMaxLogFileBackups	= 3
	deLogFileAndStdOut		= false
	defFileServerHost		= "localhost"
	defFileServerPort		= 8081
	defFileServerUser		= "admin"
	defFileServerPassword	= "admin"
	defSkipTlsVerify		= false
	defFsUseTls				= false

	flagConfigFile        	= "config-file"
	flagDbDir             	= "db-dir"
	flagServerPort        	= "server-port"
	flagLogLevel          	= "log-level"
	flagLogFile           	= "log-file"
	flagLogFileAndStdout  	= "log-file-and-stdout"
	flagLogFileMaxSize    	= "log-file-max-size"
	flagLogFileMaxBackups 	= "log-file-max-backups"
	flagLogFileMaxAge     	= "log-file-max-age"
	flagFileServerHost		= "file-server-host"
	flagFileServerPort		= "file-server-port"
	flagFileServerUser		= "file-server-user"
	flagFileServerPassword	= "file-server-password"
	flagTlsCertFile			= "tls-cert-file"
	flagTlsKeyFile			= "tls-key-file"
	flagTrustedCaFile		= "trusted-ca-file"
	flagSkipTlsVerify		= "skip-tls-verify"
	flagFsUseTls			= "fs-use-tls"
)
