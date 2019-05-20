package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/bryk-io/x/cli"
	"github.com/bryk-io/x/net/rpc"
	samplev1 "github.com/bryk-io/x/net/rpc/sample/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var buildCode = ""

var rootCmd = &cobra.Command{
	Use:   "echo-server",
	Short: "Sample echo server",
	RunE:  startServer,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	params := []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the server",
			FlagKey:   "port",
			ByDefault: 9090,
		},
		{
			Name:      "http",
			Usage:     "Enable HTTP interface",
			FlagKey:   "http",
			ByDefault: false,
		},
		{
			Name:      "cert",
			Usage:     "Certificate to use for TLS communications",
			FlagKey:   "cert",
			ByDefault: "",
		},
		{
			Name:      "key",
			Usage:     "Private key corresponding to the certificate",
			FlagKey:   "key",
			ByDefault: "",
		},
		{
			Name:      "client-ca",
			Usage:     "Certificate Authority to use when using client certificates",
			FlagKey:   "client-ca",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(rootCmd, params); err != nil {
		panic(err)
	}
}

func initConfig() {
	viper.SetEnvPrefix("echo")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func startServer(_ *cobra.Command, _ []string) (err error) {
	// Load configuration options
	log.Printf("Build code: %s\n", buildCode)
	port := viper.GetInt("port")
	enableHTTP := viper.GetBool("http")

	// Echo service provider
	echoService := &rpc.Service{
		GatewaySetup: samplev1.RegisterEchoAPIHandlerFromEndpoint,
		Setup: func(server *grpc.Server) {
			samplev1.RegisterEchoAPIServer(server, &samplev1.EchoHandler{})
		},
	}

	// Configure server
	srvOptions := []rpc.ServerOption{
		rpc.WithPort(port),
		rpc.WithService(echoService),
	}
	if enableHTTP {
		log.Println("HTTP interface enabled")
		srvOptions = append(srvOptions, rpc.WithHTTPGateway(rpc.HTTPGatewayOptions{}))
	}
	if viper.GetString("cert") != "" {
		srvTLS := rpc.ServerTLSConfig{}
		srvTLS.Cert, err = ioutil.ReadFile(viper.GetString("cert"))
		if err != nil {
			return err
		}
		srvTLS.PrivateKey, err = ioutil.ReadFile(viper.GetString("key"))
		if err != nil {
			return err
		}
		clientCA := viper.GetString("client-ca")
		if clientCA != "" {
			cc, err := ioutil.ReadFile(clientCA)
			if err != nil {
				return err
			}
			srvTLS.ClientCAs = append(srvTLS.ClientCAs, cc)
		}
		srvOptions = append(srvOptions, rpc.WithTLS(srvTLS))
	}
	server, err := rpc.NewServer(srvOptions...)
	if err != nil {
		return err
	}

	// Start server and wait for interruption signal
	go server.Start()
	log.Printf("Waiting for requests at port: %d\n", port)
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	})
	log.Println("Server closed")
	_ = server.Stop()
	return nil
}
