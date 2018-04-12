package utils

import (
	"fmt"
	"os"
	"os/signal"
)

// ShutdownRequestChannel is used to initiate shutdown from one of the
// subsystems using the same code paths as when an interrupt signal is received.
var ShutdownRequestChannel = make(chan struct{})

// InterruptSignals defines the default signals to catch in order to do a proper
// shutdown.  This may be modified during init depending on the platform.
var InterruptSignals = []os.Signal{os.Interrupt}

// InterruptListener listens for OS Signals such as SIGINT (Ctrl+C) and shutdown
// requests from shutdownRequestChannel.  It returns a channel that is closed
// when either signal is received.
func InterruptListener() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, InterruptSignals...)

		// Listen for initial shutdown signal and close the returned
		// channel to notify the caller.
		select {
		case sig := <-interruptChannel:
			fmt.Printf("Received signal (%s).  Shutting down...\n",
				sig)

		case <-ShutdownRequestChannel:
			fmt.Printf("Shutdown requested.  Shutting down...\n")
		}
		close(c)

		// Listen for repeated signals and display a message so the user
		// knows the shutdown is in progress and the process is not
		// hung.
		for {
			select {
			case sig := <-interruptChannel:
				fmt.Printf("Received signal (%s).  Already "+
					"shutting down...", sig)

			case <-ShutdownRequestChannel:
				fmt.Printf("Shutdown requested.  Already " +
					"shutting down...")
			}
		}
	}()

	return c
}
