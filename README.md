# The main service of the new Submit system. Serves as the manager of the other services in the system

### Usage:

```
submit_server

Usage:
  submit_server [command]

Available Commands:
  help        Help about any command
  start       start submit_server

Flags:
  -h, --help   help for submit_server

Use "submit_server [command] --help" for more information about a command.

```

```
start submit_server

Usage:
  submit_server start [flags]

Flags:
  -c, --config-file string            path to submit server config file
      --db-dir string                 db directory of the submit server (default "var/cache/submit-server/db")
      --file-server-host string       submit file server hostname (or ip address) (default "localhost")
      --file-server-password string   password to be used when authenticating against submit file server (default "admin")
      --file-server-port int          submit file server port (default 8081)
      --file-server-user string       user to be used when authenticating against submit file server (default "admin")
      --fs-use-tls                    use tls when accessing submit file server
  -h, --help                          help for start
      --log-file string               log to file, specify the file location
      --log-file-and-stdout           write logs to stdout if log-file is specified?
      --log-file-max-age int          maximum age of the log file before it's rotated (default 3)
      --log-file-max-backups int      maximum number of log file rotations (default 3)
      --log-file-max-size int         maximum size of the log file before it's rotated (default 10)
      --log-level string              logging level [panic, fatal, error, warn, info, debug] (default "info")
      --server-port int               port the submit server should listen on (default 8080)
      --skip-tls-verify               skip tls verification
      --tls-cert-file string          path to a file containing a certificate to use for tls
      --tls-key-file string           path to a file containing a key to use for tls
      --trusted-ca-file string        trusted ca bundle path

```