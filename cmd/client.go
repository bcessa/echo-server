package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/bryk-io/x/cli"
	"github.com/bryk-io/x/net/rpc"
	samplev1 "github.com/bryk-io/x/net/rpc/sample/v1"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start an interactive client to a running echo server",
	RunE:  runClient,
}

func init() {
	params := []cli.Param{
		{
			Name:      "tls",
			Usage:     "Enable TLS communications",
			FlagKey:   "client.tls",
			ByDefault: false,
		},
		{
			Name:      "tls-ca",
			Usage:     "Custom certificate authority to use for TLS",
			FlagKey:   "client.tls.ca",
			ByDefault: "",
		},
		{
			Name:      "client-cert",
			Usage:     "Client TLS certificate",
			FlagKey:   "client.tls.cert",
			ByDefault: "",
		},
		{
			Name:      "client-key",
			Usage:     "Client private key",
			FlagKey:   "client.tls.key",
			ByDefault: "",
		},
		{
			Name:      "rpc",
			Usage:     "RPC server endpoint",
			FlagKey:   "client.rpc",
			ByDefault: "",
		},
		{
			Name:      "http",
			Usage:     "HTTP(S) server endpoint, if any",
			FlagKey:   "client.http",
			ByDefault: "",
		},
		{
			Name:      "server-override",
			Usage:     "Server name override, must be a FQDN name, must be valid for TLS certificate if used",
			FlagKey:   "client.server",
			ByDefault: "",
		},
		{
			Name:      "auth-token",
			Usage:     "Use a dummy token as authentication mechanism",
			FlagKey:   "client.auth.token",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(clientCmd, params); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(clientCmd)
}

func getShell(cl samplev1.EchoAPIClient, hc *http.Client, endpoint string) *ishell.Shell {
	shell := ishell.New()
	shell.AddCmd(&ishell.Cmd{
		Name: "ping",
		Help: "Send a reachability test to the server",
		Func: func(c *ishell.Context) {
			if r, err := cl.Ping(context.TODO(), &types.Empty{}); err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Printf("status: %v\n", r.Ok)
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "health",
		Help: "Send a state check to the server",
		Func: func(c *ishell.Context) {
			if r, err := cl.Health(context.TODO(), &types.Empty{}); err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Printf("alive: %v\n", r.Alive)
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "request",
		Help: "Perform an 'Echo' request",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("you must specify the contents of the request")
				return
			}
			if r, err := cl.Request(context.TODO(), &samplev1.EchoRequest{Value:c.Args[0]}); err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Printf("%v\n", r.Result)
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "faulty",
		Help: "Run a faulty request, should return an error about 20% of the time",
		Func: func(c *ishell.Context) {
			if _, err := cl.Faulty(context.TODO(), &types.Empty{}); err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Println("ok")
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "slow",
		Help: "Run a slow request, should exhibit a latency between 10 to 200ms",
		Func: func(c *ishell.Context) {
			start := time.Now()
			if _, err := cl.Slow(context.TODO(), &types.Empty{}); err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Printf("latency: %dms\n", int64(time.Since(start) / time.Millisecond))
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "stress",
		Help: "Perform a simple stress test against the server. Specify method and number of requests.",
		Func: func(c *ishell.Context) {
			m := "slow"
			r := 10
			var err error
			if len(c.Args) == 2 {
				m = c.Args[0]
				r, err = strconv.Atoi(c.Args[1])
				if err != nil {
					c.Println("Your second parameter must be an integer number")
					return
				}
			}

			val := 0
			switch m {
			case "slow":
				c.ProgressBar().Final("done ")
				c.ProgressBar().Start()
				for i := 1; i <= r; i++ {
					val = (i * 100) / r
					c.ProgressBar().Suffix(fmt.Sprint(" ", val, "%"))
					c.ProgressBar().Progress(val)
					if _, err = cl.Slow(context.TODO(), &types.Empty{}); err != nil {
						c.Printf("error: %s\n", err.Error())
					}
				}
				c.ProgressBar().Stop()
			case "faulty":
				errCount := 0
				for i := 0; i < r; i++ {
					_, err = cl.Faulty(context.TODO(), &types.Empty{})
					if err != nil {
						errCount++
					}
				}
				c.Printf("error rate: %d%%\n", (errCount * 100) / r)
			default:
				c.Println("invalid method name")
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "http",
		Help: "Send an http request",
		Func: func(c *ishell.Context) {
			if endpoint == "" {
				c.Println("no HTTP endpoint specified")
				return
			}
			r, err := hc.Post(endpoint + "/echo/ping", "application/json", nil)
			if err != nil {
				c.Printf("error: %s\n", err.Error())
			} else {
				c.Printf("Status: %v\n", r.Status)
				c.Println("--------------")
				for k, v := range r.Header {
					c.Printf("%s: %v\n", k, v)
				}
				_ = r.Body.Close()
				return
			}
		},
	})
	return shell
}

func getHTTPClient(ca []byte, cert *tls.Certificate) *http.Client {
	cl := &http.Client{}
	if ca != nil {
		cp, _ := x509.SystemCertPool()
		cp.AppendCertsFromPEM(ca)
		conf := &tls.Config{RootCAs: cp}
		if cert != nil {
			conf.Certificates = []tls.Certificate{*cert}
		}
		cl.Transport = &http.Transport{TLSClientConfig: conf}
	}
	return cl
}

func runClient(_ *cobra.Command, _ []string) error {
	var clientCert *tls.Certificate
	var clientCA []byte = nil
	authToken := viper.GetString("client.auth.token")
	endpoint := viper.GetString("client.rpc")
	if endpoint == "" {
		return errors.New("you must specify the RPC endpoint")
	}

	// Base client configuration
	clOpts := []rpc.ClientOption{
		rpc.WaitForReady(),
		rpc.WithTimeout(5 * time.Second),
		rpc.WithCompression(),
		rpc.WithUserAgent("echo-client/0.1.0"),
	}

	// Server name override, should only be used while testing
	if viper.GetString("client.server") != "" {
		clOpts = append(clOpts, rpc.WithServerNameOverride("client.server"))
	}

	// Authentication by token
	if authToken != "" {
		log.Printf("authenticating with token: %s\n", authToken)
		clOpts = append(clOpts, rpc.WithAuthToken(authToken))
	}

	// Authentication by certificate
	if viper.GetString("client.tls.cert") != "" {
		if !viper.GetBool("client.tls") {
			return errors.New("client certificates can only be used with TLS connections")
		}
		log.Println("authenticating with client certificate")
		cert, err := ioutil.ReadFile(viper.GetString("client.tls.cert"))
		if err != nil {
			return err
		}
		key, err := ioutil.ReadFile(viper.GetString("client.tls.key"))
		if err != nil {
			return err
		}
		clOpts = append(clOpts, rpc.WithAuthCertificate(cert, key))

		// Load client certificate for HTTP client if required
		if viper.GetString("client.http") != "" {
			log.Println("loading client certificate for HTTP client")
			hc, err := rpc.LoadCertificate(cert, key)
			if err != nil {
				return err
			}
			clientCert = &hc
		}
	}

	// TLS setup
	if viper.GetBool("client.tls") {
		log.Println("TLS enabled")
		var err error
		clientTLS := rpc.ClientTLSConfig{
			IncludeSystemCAs: true,
		}
		// Load custom CA, if any
		if viper.GetString("client.tls.ca") != "" {
			log.Printf("custom certificate authority: %s\n", viper.GetString("client.tls.ca"))
			clientCA, err = ioutil.ReadFile(viper.GetString("client.tls.ca"))
			if err != nil {
				return err
			}
			clientTLS.CustomCAs = append(clientTLS.CustomCAs, clientCA)
		}
		clOpts = append(clOpts, rpc.WithClientTLS(clientTLS))
	}

	// Open connection
	log.Printf("reaching out to: %s\n", endpoint)
	conn, err := rpc.NewClientConnection(endpoint, clOpts...)
	if err != nil {
		return err
	}
	log.Println("connection ready")

	// Start interactive client
	cl := samplev1.NewEchoAPIClient(conn)
	shell := getShell(cl, getHTTPClient(clientCA, clientCert), viper.GetString("client.http"))
	shell.Println("=== interactive shell")
	shell.Run()

	// Close connection
	log.Println("closing client")
	return conn.Close()
}
