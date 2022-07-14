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

func runClient(ctx context.Context, host string, token string, tag string, debug bool) error {
	n := new(Netbox)
	n1 := new(Netbox)
	// token := "0123456789abcdef0123456789abcdef01234567"
	// host := "localhost:8000"

	n.logger = defaultLogger("debug")
	n1.logger = defaultLogger("debug")
	n.debug = debug
	n1.debug = debug

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("%v", err)
			}
		default:
			err := n.ReadFromNetbox(ctx, host, token)
			time.Sleep(time.Second)
			if err != nil {
				return fmt.Errorf("read from Netbox failed: %v", err)
			}
			err = n1.ReadFromNetboxFiltered(ctx, host, token, tag)
			if err != nil {
				return fmt.Errorf("filtered Read from Netbox failed: %v", err)
			}
			time.Sleep(time.Second)
			ret, err2 := n1.SerializeMachines(n1.Records)
			if err2 != nil {
				return fmt.Errorf("error serializing machines: %v", err2)
			}
			machines, err3 := ReadMachinesBytes(ctx, ret, n1)
			if err3 != nil {
				return fmt.Errorf("error reading Bytes: %v", err3)
			}

			_, err = WriteToCsv(ctx, machines, n1)
			if err != nil {
				return fmt.Errorf("error writing to csv: %v", err)
			}

		}
		return nil
	}

	// err = n1.ReadFromNetboxFiltered(ctx, host, token, tag)
	// if err != nil {
	// 	return fmt.Errorf("filtered Read from Netbox failed: %v", err)
	// }

}

// defaultLogger is a zerolog logr implementation.
func defaultLogger(level string) logr.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerologr.NameFieldName = "logger"
	zerologr.NameSeparator = "/"

	zl := zerolog.New(os.Stdout)
	zl = zl.With().Caller().Timestamp().Logger()
	var l zerolog.Level
	switch level {
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}
	zl = zl.Level(l)

	return zerologr.New(&zl)
}
