module github.com/bcessa/echo-server

go 1.12

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/bryk-io/x v0.0.0-20190611172044-70298fee2853
	github.com/fatih/color v1.7.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	google.golang.org/grpc v1.21.0
)

replace (
	github.com/cloudflare/cfssl => github.com/bryk-io/cfssl v0.0.0-20190303174050-7d50b68e4142
	github.com/dgraph-io/badger v1.5.5 => github.com/bryk-io/badger v1.5.5
	github.com/grpc-ecosystem/go-grpc-middleware => github.com/bryk-io/go-grpc-middleware v1.0.1-0.20190419153159-d28668ee9f4e
)
