module github.com/bcessa/echo-server

go 1.13

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	go.bryk.io/x v0.0.0-20191206191545-f9a10f6a12ad
	google.golang.org/grpc v1.23.0
)

replace github.com/cloudflare/cfssl => github.com/bryk-io/cfssl v0.0.0-20190303174050-7d50b68e4142
