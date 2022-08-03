package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

// runClient Orchestrates all the business logic and calls relevant functions to return the csv file.
func runClient(ctx context.Context, host string, token string, tag string, debug bool) error {
	n := new(Netbox)
<<<<<<< HEAD

=======
>>>>>>> 9c3512e4 (Unexported functions, reused err variable in client.go)
	n.logger = defaultLogger(debug)

	err := n.readFromNetboxFiltered(ctx, host, token, tag)
	if err != nil {
		return fmt.Errorf("filtered Read from Netbox failed: %v", err)
	}
	time.Sleep(time.Second)
	ret, err := n.serializeMachines(n.Records)
	if err != nil {
		return fmt.Errorf("error serializing machines: %v", err)
	}
<<<<<<< HEAD
	machines, err3 := ReadMachinesBytes(ret, n)
	if err3 != nil {
		return fmt.Errorf("error reading Bytes: %v", err3)
=======
	machines, err := readMachinesBytes(ctx, ret, n)
	if err != nil {
		return fmt.Errorf("error reading Bytes: %v", err)
>>>>>>> 9c3512e4 (Unexported functions, reused err variable in client.go)
	}
	n.logger.Info("All API calls done")
	time.Sleep(time.Second)
	err = writeToCSVHelper(ctx, machines, n)
	if err != nil {
		return fmt.Errorf("error writing to csv: %v", err)
	}
	return nil
}

// defaultLogger is a zerolog logr implementation.
func defaultLogger(debug bool) logr.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerologr.NameFieldName = "logger"
	zerologr.NameSeparator = "/"

	zl := zerolog.New(os.Stdout)
	zl = zl.With().Caller().Timestamp().Logger()
	var l zerolog.Level

	if debug {
		l = zerolog.DebugLevel
	} else {
		l = zerolog.InfoLevel
	}
	zl = zl.Level(l)

	return zerologr.New(&zl)
}
