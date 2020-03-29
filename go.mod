module github.com/bcessa/echo-server

go 1.13

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	go.bryk.io/x v0.0.0-20200304202726-6fc2d592300e
	google.golang.org/grpc v1.27.1
)

replace github.com/cloudflare/cfssl => github.com/bryk-io/cfssl v0.0.0-20190303174050-7d50b68e4142
