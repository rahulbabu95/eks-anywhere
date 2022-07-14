package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// textctl is a simple applications in which all commands are built up in func
// main. It demonstrates how to declare minimal commands, how to wire them
// together into a command tree, and one way to allow subcommands access to
// flags set in parent commands.

func main() {

	// keep ffcli or remove .? --done
	// Clean up the stdout -- done
	// log level. -- flag for logging.
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer done()

	host := flag.String("host", "", "Host Server running DCIM tool")
	token := flag.String("token", "", "API token for HTTP connection with DCIM")
	tag := flag.String("tag", "eks-a", "tag for filtering devices")
	debug := flag.Bool("debug", false, "debug flag for logging")
	flag.Parse()
	if len(*host) == 0 {
		fmt.Fprintln(os.Stdout, "Host cannot be blank")
	} else if len(*token) == 0 {
		fmt.Fprintln(os.Stdout, "token ID cannot be blank")
	} else if *debug {
		fmt.Println("----------------------------DEBUG LOGS------------------------------------")
		err := runClient(ctx, *host, *token, *tag, *debug)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		err := runClient(ctx, *host, *token, *tag, *debug)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
