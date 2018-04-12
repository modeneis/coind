// a local fake skyd that reports deposits to teller for testing
package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/modeneis/coind/src/server/api"
	"github.com/modeneis/coind/src/server/rpc"
	"github.com/modeneis/coind/src/server/utils"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	//flags
	keyFile := flag.String("key", "rpc.key", "btcd rpc key")
	certFile := flag.String("cert", "rpc.cert", "btcd rpc cert")
	address := flag.String("address", "127.0.0.1:8334", "btcd listening address")
	httpAPIAddress := flag.String("api", "127.0.0.1:4122", "http api listening address")

	flag.Parse()

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interruptedChan := utils.InterruptListener()

	srv := rpc.RpcServer{
		RequestProcessShutdown: make(chan struct{}),
		Key:               *keyFile,
		Cert:              *certFile,
		Address:           *address,
		MaxConcurrentReqs: 10,
	}

	apiServer := api.NewHTTPAPIServer(*httpAPIAddress)

	defer func() {
		apiServer.Stop()
		if err := srv.Stop(); err != nil {
			fmt.Println("server.Stop failed:", err)
		}
		fmt.Printf("Shutdown complete\n")
	}()

	// Signal process shutdown when the RPC server requests it.
	go func() {
		<-srv.RequestedProcessShutdown()
		utils.ShutdownRequestChannel <- struct{}{}
	}()

	srv.Start()

	go func() {
		fmt.Printf("HTTP API server listening on http://%s\n", *httpAPIAddress)
		err := apiServer.Start()
		if err != nil {
			fmt.Printf("HTTP API server failed to start\n")
		}
	}()

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interruptedChan
	return nil
}
