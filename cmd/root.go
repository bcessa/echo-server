package cmd

import (
	"fmt"
	"io/ioutil"
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
	// Add support for ECHO_ env variables
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("echo")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
	})

	// Define server parameters
	params := []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the main RPC server",
			FlagKey:   "port",
			ByDefault: 9090,
		},
		{
			Name:      "name",
			Usage:     "FQDN server name, must be valid for TLS certificate if used",
			FlagKey:   "name",
			ByDefault: "localhost",
		},
		{
			Name:      "tls-cert",
			Usage:     "Certificate to use for TLS communications",
			FlagKey:   "tls-cert",
			ByDefault: "",
		},
		{
			Name:      "tls-key",
			Usage:     "Private key corresponding to the TLS certificate",
			FlagKey:   "tls-key",
			ByDefault: "",
		},
		{
			Name:      "tls-ca",
			Usage:     "Custom Certificate Authority to use for TLS communications",
			FlagKey:   "tls-ca",
			ByDefault: "",
		},
		{
			Name:      "client-ca",
			Usage:     "Custom Certificate Authority to use for client authentication",
			FlagKey:   "client-ca",
			ByDefault: "",
		},
		{
			Name:      "client-cert",
			Usage:     "Require clients to present identity certificates",
			FlagKey:   "client-cert",
			ByDefault: false,
		},
		{
			Name:      "http",
			Usage:     "Enable HTTP interface",
			FlagKey:   "http",
			ByDefault: false,
		},
		{
			Name:      "http-gw-port",
			Usage:     "Port to use for the HTTP gateway interface",
			FlagKey:   "http-gw-port",
			ByDefault: 9091,
		},
		{
			Name:      "http-gw-cert",
			Usage:     "Client certificate used by the HTTP gateway component",
			FlagKey:   "http-gw-cert",
			ByDefault: "",
		},
		{
			Name:      "http-gw-key",
			Usage:     "Private key used by the HTTP gateway component",
			FlagKey:   "http-gw-key",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(rootCmd, params); err != nil {
		panic(err)
	}
}

func startServer(_ *cobra.Command, _ []string) (err error) {
	// Load configuration options
	fmt.Printf("= build code: %s\n", buildCode)
	port := viper.GetInt("port")
	var cert []byte
	var key []byte

	// Echo service provider
	echoService := &rpc.Service{
		GatewaySetup: samplev1.RegisterEchoAPIHandlerFromEndpoint,
		Setup: func(server *grpc.Server) {
			samplev1.RegisterEchoAPIServer(server, &samplev1.EchoHandler{})
		},
	}

	// Base server configuration
	srvOptions := []rpc.ServerOption{
		rpc.WithNetworkInterface(rpc.NetworkInterfaceAll),
		rpc.WithServerName(viper.GetString("name")),
		rpc.WithPort(port),
		rpc.WithService(echoService),
		rpc.WithLogger(nil),
		rpc.WithPanicRecovery(),
	}

	// TLS configuration
	if viper.GetString("tls-cert") != "" {
		fmt.Println("= TLS enabled")
		var err error
		srvTLS := rpc.ServerTLSConfig{IncludeSystemCAs: true}
		if viper.GetString("tls-ca") != "" {
			ca, err := ioutil.ReadFile(viper.GetString("tls-ca"))
			if err != nil {
				return err
			}
			srvTLS.CustomCAs = append(srvTLS.CustomCAs, ca)
		}
		srvTLS.Cert, err = ioutil.ReadFile(viper.GetString("tls-cert"))
		if err != nil {
			return err
		}
		srvTLS.PrivateKey, err = ioutil.ReadFile(viper.GetString("tls-key"))
		if err != nil {
			return err
		}
		srvOptions = append(srvOptions, rpc.WithTLS(srvTLS))
		cert = srvTLS.Cert
		key = srvTLS.PrivateKey
	}

	// Cert-based authentication configuration
	if viper.GetBool("client-cert") {
		fmt.Println("= enabling certificate-based authentication")
		clientCA, err := ioutil.ReadFile(viper.GetString("client-ca"))
		if err != nil {
			return err
		}
		srvOptions = append(srvOptions, rpc.WithAuthByCert(clientCA))
	}

	// HTTP gateway configuration
	if viper.GetBool("http") {
		fmt.Printf("= HTTP interface enabled on port: %d\n", viper.GetInt("http-gw-port"))
		var err error
		gwOpts := rpc.HTTPGatewayOptions{
			Port: viper.GetInt("http-port"),
		}
		// Load custom gateway client cert if provided
		if viper.GetString("http-gw-cert") != "" {
			cert, err = ioutil.ReadFile(viper.GetString("http-gw-cert"))
			if err != nil {
				return err
			}
			key, err = ioutil.ReadFile(viper.GetString("http-gw-key"))
			if err != nil {
				return err
			}
		}
		if len(cert) > 0 {
			gwOpts.ClientCertificate = cert
			gwOpts.ClientPrivateKey = key
		}
		srvOptions = append(srvOptions, rpc.WithHTTPGateway(gwOpts))
	}

	// Start server and wait for interruption signal
	server, err := rpc.NewServer(srvOptions...)
	if err != nil {
		return err
	}
	go server.Start()
	fmt.Printf("= waiting for requests at port: %d\n", port)
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	})
	fmt.Println("= server closed")
	_ = server.Stop()
	return nil
}
