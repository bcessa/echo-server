module github.com/bcessa/echo-server

go 1.12

require (
	github.com/bryk-io/x v0.0.0-20190518151150-eab10ec0beb5
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	google.golang.org/grpc v1.21.0
)

replace (
	github.com/cloudflare/cfssl => github.com/bryk-io/cfssl v0.0.0-20190303174050-7d50b68e4142
	github.com/dgraph-io/badger v1.5.5 => github.com/bryk-io/badger v1.5.5
	github.com/grpc-ecosystem/go-grpc-middleware => github.com/bryk-io/go-grpc-middleware v1.0.1-0.20190419153159-d28668ee9f4e
)
