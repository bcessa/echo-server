package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"go.bryk.io/x/cli"
	"go.bryk.io/x/net/rpc"
	samplev1 "go.bryk.io/x/net/rpc/sample/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
			Name:      "monitoring",
			Usage:     "Produce metrics that can be consumed by instrumentation services (requires HTTP)",
			FlagKey:   "server.monitoring",
			ByDefault: false,
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
		{
			Name:      "log-json",
			Usage:     "Log messages in JSON format (text by default)",
			FlagKey:   "server.log.json",
			ByDefault: false,
		},
	}
	if err := cli.SetupCommandParams(serverCmd, params); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(serverCmd)
}

func startServer(_ *cobra.Command, _ []string) (err error) {
	// Load configuration options
	port := viper.GetInt("server.port")

	// Echo service provider
	echoService := &rpc.Service{
		GatewaySetup: samplev1.RegisterEchoAPIHandlerFromEndpoint,
		ServerSetup: func(server *grpc.Server) {
			samplev1.RegisterEchoAPIServer(server, &samplev1.EchoHandler{})
		},
	}

	// TLS custom CA
	var tlsCA []byte = nil

	// Logger
	ll := logrus.New()
	if viper.GetBool("server.log.json") {
		ll.SetFormatter(new(logrus.JSONFormatter))
	} else {
		formatter := &prefixed.TextFormatter{
			FullTimestamp: true,
			TimestampFormat: time.StampMilli,
		}
		formatter.SetColorScheme(&prefixed.ColorScheme{
			DebugLevelStyle: "black",
			TimestampStyle:  "white+h",
		})
		ll.SetFormatter(formatter)
	}
	le := logrus.NewEntry(ll)

	// Base server configuration
	srvOptions := []rpc.ServerOption{
		rpc.WithNetworkInterface(rpc.NetworkInterfaceAll),
		rpc.WithPort(port),
		rpc.WithService(echoService),
		rpc.WithInputValidation(),
		rpc.WithPanicRecovery(),
		rpc.WithLogger(rpc.LoggingOptions{
			Mode:   rpc.LOGRUS,
			Logrus: le,
			FilterMethods: []string{
				"bryk.x.net.rpc.sample.v1.EchoAPI/Ping",
			},
		}),
	}

	// Authentication by token
	if token := viper.GetString("server.auth.token"); token != "" {
		le.Infof("enabling token validation with dummy value: %s", token)
		tv := rpc.WithAuthByToken(func(t string) (code codes.Code, s string) {
			if token != t {
				return codes.Unauthenticated, fmt.Sprintf("invalid token provided '%s'", t)
			}
			return codes.OK, ""
		})
		srvOptions = append(srvOptions, tv)
	}

	// Authentication by certificate
	if clientCA := viper.GetString("server.auth.ca"); clientCA != "" {
		le.Infof("enabling certificate-based authentication: %s", clientCA)
		ca, err := ioutil.ReadFile(clientCA)
		if err != nil {
			return err
		}
		srvOptions = append(srvOptions, rpc.WithAuthByCertificate(ca))
	}

	// TLS configuration
	if viper.GetString("server.tls.cert") != "" {
		le.Info("TLS enabled")
		var err error
		srvTLS := rpc.ServerTLSConfig{IncludeSystemCAs: true}
		le.Debugf("loading certificate: %s", viper.GetString("server.tls.cert"))
		srvTLS.Cert, err = ioutil.ReadFile(viper.GetString("server.tls.cert"))
		if err != nil {
			return err
		}
		le.Debugf("loading private key: %s", viper.GetString("server.tls.key"))
		srvTLS.PrivateKey, err = ioutil.ReadFile(viper.GetString("server.tls.key"))
		if err != nil {
			return err
		}

		// Load custom CA if used
		if viper.GetString("server.tls.ca") != "" {
			le.Debugf("loading CA: %s", viper.GetString("server.tls.ca"))
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
		le.Infof("HTTP interface enabled on port: %d", viper.GetInt("server.http.port"))

		// Gateway internal client options
		gwCl := []rpc.ClientOption{
			// Internal connection from HTTP proxy to RPC server takes any provided certificate as valid
			rpc.WithInsecureSkipVerify(),
		}

		// Server is using TLS
		if viper.GetString("server.tls.cert") != "" {
			gwTLS := rpc.ClientTLSConfig{IncludeSystemCAs: true}
			if tlsCA != nil {
				gwTLS.CustomCAs = append(gwTLS.CustomCAs, tlsCA)
			}
			gwCl = append(gwCl, rpc.WithClientTLS(gwTLS))
		}

		// Load custom gateway client cert if provided
		if viper.GetString("server.http.cert") != "" {
			le.Debugf("gateway client certificate: %s", viper.GetString("server.http.cert"))
			cert, err := ioutil.ReadFile(viper.GetString("server.http.cert"))
			if err != nil {
				return err
			}
			le.Debugf("gateway private key: %s", viper.GetString("server.http.key"))
			key, err := ioutil.ReadFile(viper.GetString("server.http.key"))
			if err != nil {
				return err
			}
			gwCl = append(gwCl, rpc.WithAuthCertificate(cert, key))
		}

		// Get gateway instance
		gwOpts := []rpc.HTTPGatewayOption{
			rpc.WithGatewayPort(viper.GetInt("server.http.port")),
			rpc.WithClientOptions(gwCl),
		}
		gw, err := rpc.NewHTTPGateway(gwOpts...)
		if err != nil {
			return err
		}
		srvOptions = append(srvOptions, rpc.WithHTTPGateway(gw))

		// Enable monitoring
		if viper.GetBool("server.monitoring") {
			// Locally run dev instances of prometheus and grafana for testing.
			// docker run -d --rm --name prometheus -p 4000:9090 -v monitor.yaml:/etc/prometheus/prometheus.yml prom/prometheus
			// docker run -d --rm --name grafana -p 3000:3000 -e "GF_SECURITY_ADMIN_PASSWORD=password" grafana/grafana
			le.Info("monitoring enabled on endpoint: /metrics")
			srvOptions = append(srvOptions, rpc.WithMonitoring(rpc.MonitoringOptions{
				IncludeHistograms:   true,
				UseGoCollector:      true,
				UseProcessCollector: true,
			}))
		}
	}

	// Start server
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

	// Wait for server to be ready
	<-ready
	le.Infof("waiting for requests at port: %d", port)

	// Catch interruption signals and quit
	<-cli.SignalsHandler([]os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	})
	le.Warn("server closed")
 	_ = server.Stop(true)
	return nil
}
