package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
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
	Use:   "client {SERVER}",
	Short: "Start an interactive client to a running echo server",
	RunE:  runClient,
}

func init() {
	params := []cli.Param{
		{
			Name:      "ca",
			Usage:     "Certificate Authority to use",
			FlagKey:   "client.ca",
			ByDefault: "",
		},
		{
			Name:      "cert",
			Usage:     "Client TLS certificate",
			FlagKey:   "client.cert",
			ByDefault: "",
		},
		{
			Name:      "key",
			Usage:     "Client private key",
			FlagKey:   "client.key",
			ByDefault: "",
		},
	}
	if err := cli.SetupCommandParams(clientCmd, params); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(clientCmd)
}

func getShell(cl samplev1.EchoAPIClient) *ishell.Shell {
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
				c.Printf("latency: %dms\n", int64(time.Now().Sub(start) / time.Millisecond))
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
	return shell
}

func runClient(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("you must specify the server endpoint")
	}
	endpoint := args[0]

	// Configure client connection
	clOpts := []rpc.ClientOption{
		rpc.WaitForReady(),
		rpc.WithTimeout(5 * time.Second),
		rpc.WithCompression(),
		rpc.WithUserAgent("echo-client/0.1.0"),
	}
	if viper.GetString("client.cert") != "" {
		fmt.Println("= TLS enabled")
		var err error
		server := strings.Split(endpoint, ":")
		clientTLS := rpc.ClientTLSConfig{
			ServerName:       server[0],
			IncludeSystemCAs: true,
			CustomCACerts:    [][]byte{},
		}
		if clientTLS.ClientCertificate, err = ioutil.ReadFile(viper.GetString("client.cert")); err != nil {
			return err
		}
		if clientTLS.ClientPrivateKey, err = ioutil.ReadFile(viper.GetString("client.key")); err != nil {
			return err
		}
		ca, err := ioutil.ReadFile(viper.GetString("client.ca"))
		if err != nil {
			return err
		}
		clientTLS.CustomCACerts = append(clientTLS.CustomCACerts, ca)
		clOpts = append(clOpts, rpc.WithClientTLS(clientTLS))
	}

	// Open connection
	fmt.Printf("= reaching out to: %s\n", endpoint)
	conn, err := rpc.NewClientConnection(endpoint, clOpts...)
	if err != nil {
		return err
	}
	fmt.Println("= connection ready")

	// Start interactive client
	cl := samplev1.NewEchoAPIClient(conn)
	shell := getShell(cl)
	shell.Println("= interactive shell")
	shell.Run()

	// Close connection
	fmt.Println("= closing client")
	return conn.Close()
}
