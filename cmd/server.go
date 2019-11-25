package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/bryk-io/x/cli"
	"github.com/bryk-io/x/net/rpc"
	samplev1 "github.com/bryk-io/x/net/rpc/sample/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start echo server",
	RunE:  startServer,
}

func init() {
	// Define server parameters
	params := []cli.Param{
		{
			Name:      "port",
			Usage:     "TCP port to use for the main RPC server",
			FlagKey:   "server.port",
			ByDefault: 9090,
		},
		{
			Name:      "tls-cert",
			Usage:     "Certificate to use for TLS communications",
			FlagKey:   "server.tls.cert",
			ByDefault: "",
		},
		{
			Name:      "tls-key",
			Usage:     "Private key corresponding to the TLS certificate",
			FlagKey:   "server.tls.key",
			ByDefault: "",
		},
		{
			Name:      "tls-ca",
			Usage:     "Custom Certificate Authority to use for TLS communications",
			FlagKey:   "server.tls.ca",
			ByDefault: "",
		},
		{
			Name:      "http",
			Usage:     "Enable HTTP interface",
			FlagKey:   "server.http",
			ByDefault: false,
		},
		{
			Name:      "http-port",
			Usage:     "Port to use for the HTTP gateway interface",
			FlagKey:   "server.http.port",
			ByDefault: 9091,
		},
		{
			Name:      "http-cert",
			Usage:     "Client certificate used by the HTTP gateway component",
			FlagKey:   "server.http.cert",
			ByDefault: "",
		},
		{
			Name:      "http-key",
			Usage:     "Private key used by the HTTP gateway component",
			FlagKey:   "server.http.key",
			ByDefault: "",
		},
		{
			Name:      "auth-by-token",
			Usage:     "Used a dummy token for authentication",
			FlagKey:   "server.auth.token",
			ByDefault: "",
		},
		{
			Name:      "auth-by-cert",
			Usage:     "Provide the CA used to verify client certificates as credentials",
			FlagKey:   "server.auth.ca",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(serverCmd, params); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(serverCmd)
}

func startServer(_ *cobra.Command, _ []string) (err error) {
	// Load configuration options
	log.Printf("build code: %s\n", buildCode)
	port := viper.GetInt("server.port")

	// Echo service provider
	echoService := &rpc.Service{
		GatewaySetup: samplev1.RegisterEchoAPIHandlerFromEndpoint,
		Setup: func(server *grpc.Server) {
			samplev1.RegisterEchoAPIServer(server, &samplev1.EchoHandler{})
		},
	}

	// TLS custom CA
	var tlsCA []byte = nil

	// Base server configuration
	srvOptions := []rpc.ServerOption{
		rpc.WithNetworkInterface(rpc.NetworkInterfaceAll),
		rpc.WithPort(port),
		rpc.WithService(echoService),
		rpc.WithLogger(nil),
		rpc.WithPanicRecovery(),
	}

	// Authentication by token
	if token := viper.GetString("server.auth.token"); token != "" {
		log.Printf("enabling token validation with dummy value: %s\n", token)
		tv := rpc.WithAuthByTokenValidator(func(t string) bool {
			return token == t
		})
		srvOptions = append(srvOptions, tv)
	}

	// Authentication by certificate
	if clientCA := viper.GetString("server.auth.ca"); clientCA != "" {
		log.Printf("enabling certificate-based authentication: %s\n", clientCA)
		ca, err := ioutil.ReadFile(clientCA)
		if err != nil {
			return err
		}
		srvOptions = append(srvOptions, rpc.WithAuthByCertificate(ca))
	}

	// TLS configuration
	if viper.GetString("server.tls.cert") != "" {
		log.Println("TLS enabled")
		var err error
		srvTLS := rpc.ServerTLSConfig{IncludeSystemCAs: true}
		log.Printf("loading certifiate: %s\n", viper.GetString("server.tls.cert"))
		srvTLS.Cert, err = ioutil.ReadFile(viper.GetString("server.tls.cert"))
		if err != nil {
			return err
		}
		log.Printf("loading private key: %s\n", viper.GetString("server.tls.key"))
		srvTLS.PrivateKey, err = ioutil.ReadFile(viper.GetString("server.tls.key"))
		if err != nil {
			return err
		}

		// Load custom CA if used
		if viper.GetString("server.tls.ca") != "" {
			log.Printf("loading CA: %s\n", viper.GetString("server.tls.ca"))
			tlsCA, err = ioutil.ReadFile(viper.GetString("server.tls.ca"))
			if err != nil {
				return err
			}
			srvTLS.CustomCAs = append(srvTLS.CustomCAs, tlsCA)
		}
		srvOptions = append(srvOptions, rpc.WithTLS(srvTLS))
	}

	// HTTP gateway configuration
	if viper.GetBool("server.http") {
		log.Printf("HTTP interface enabled on port: %d\n", viper.GetInt("server.http.port"))
		gwOpts := rpc.HTTPGatewayOptions{
			Port: viper.GetInt("server.http.port"),
			ClientOptions: []rpc.ClientOption{
				// Internal connection from HTTP proxy to RPC server takes any provided certificate as valid
				rpc.WithInsecureSkipVerify(),
			},
		}

		// Server is using TLS
		if viper.GetString("server.tls.cert") != "" {
			gwTLS := rpc.ClientTLSConfig{IncludeSystemCAs: true}
			if tlsCA != nil {
				gwTLS.CustomCAs = append(gwTLS.CustomCAs, tlsCA)
			}
			gwOpts.ClientOptions = append(gwOpts.ClientOptions, rpc.WithClientTLS(gwTLS))
		}

		// Load custom gateway client cert if provided
		if viper.GetString("server.http.cert") != "" {
			log.Printf("gateway client certificate: %s\n", viper.GetString("server.http.cert"))
			cert, err := ioutil.ReadFile(viper.GetString("server.http.cert"))
			if err != nil {
				return err
			}
			log.Printf("gateway private key: %s\n", viper.GetString("server.http.key"))
			key, err := ioutil.ReadFile(viper.GetString("server.http.key"))
			if err != nil {
				return err
			}
			gwOpts.ClientOptions = append(gwOpts.ClientOptions, rpc.WithAuthCertificate(cert, key))
		}
		srvOptions = append(srvOptions, rpc.WithHTTPGateway(gwOpts))
	}

	// Start server and wait for interruption signal
	ready := make(chan bool)
	server, err := rpc.NewServer(srvOptions...)
	if err != nil {
		return err
	}
	go func() {
		if err := server.Start(ready); err != nil {
			log.Println("failed to start server:", err)
		}
	}()
	<-ready
	log.Printf("waiting for requests at port: %d\n", port)
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	})
	log.Println("server closed")
	_ = server.Stop()
	return nil
}
