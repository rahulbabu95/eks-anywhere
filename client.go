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
	// token := "0123456789abcdef0123456789abcdef01234567"
	// host := "localhost:8000"

	n.logger = defaultLogger(debug)

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("%v", err)
			}
		default:
			if err := n.ReadFromNetboxFiltered(ctx, host, token, tag); err != nil {
				return fmt.Errorf("filtered Read from Netbox failed: %v", err)
			}
			time.Sleep(time.Second)
			ret, err2 := n.SerializeMachines(n.Records)
			if err2 != nil {
				return fmt.Errorf("error serializing machines: %v", err2)
			}
			machines, err3 := ReadMachinesBytes(ctx, ret, n)
			if err3 != nil {
				return fmt.Errorf("error reading Bytes: %v", err3)
			}

			_, err := WriteToCsv(ctx, machines, n)
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
