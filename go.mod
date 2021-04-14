module github.com/DAv10195/submit_server

go 1.13

require (
	github.com/DAv10195/submit_commons v0.0.0-20210414053531-8066a0155d69
	github.com/boltdb/bolt v1.3.1
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	golang.org/x/sys v0.0.0-20201221093633-bc327ba9c2f0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/DAv10195/submit_commons => ../submit_commons
