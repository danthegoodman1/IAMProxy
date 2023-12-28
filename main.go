package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/control_plane"
	"github.com/danthegoodman1/GoAPITemplate/observability"
	"github.com/danthegoodman1/GoAPITemplate/tracing"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danthegoodman1/GoAPITemplate/gologger"
	"github.com/danthegoodman1/GoAPITemplate/http_server"
	"github.com/danthegoodman1/GoAPITemplate/utils"
)

var logger = gologger.NewLogger()

func main() {
	mainCtx := context.Background()
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			logger.Error().Err(err).Msg("error loading .env file, exiting")
			os.Exit(1)
		}
	}
	logger.Debug().Msg("starting IAM Proxy api")

	prometheusReporter := observability.NewPrometheusReporter()
	go func() {
		err := observability.StartInternalHTTPServer(":8042", prometheusReporter)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("internal server couldn't start")
			os.Exit(1)
		}
	}()

	g := errgroup.Group{}
	if utils.CacheEnabled {
		g.Go(func() error {
			logger.Debug().Msg("starting groupcache")
			control_plane.InitCache(mainCtx)
			return nil
		})
	}

	var tracer *trace.TracerProvider
	if utils.Env_TracingEnabled {
		logger.Info().Msg("enabling tracing")
		g.Go(func() (err error) {
			tracer, err = tracing.InitTracer(mainCtx)
			return
		})
	}

	err := g.Wait()
	if err != nil {
		logger.Error().Err(err).Msg("Error starting services")
	}

	httpServer := http_server.StartHTTPServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")

	// For AWS ALB needing some time to de-register pod
	// Convert the time to seconds
	sleepTime := utils.GetEnvOrDefaultInt("SHUTDOWN_SLEEP_SEC", 0)
	logger.Info().Msg(fmt.Sprintf("sleeping for %ds before exiting", sleepTime))

	time.Sleep(time.Second * time.Duration(sleepTime))
	logger.Info().Msg(fmt.Sprintf("slept for %ds, exiting", sleepTime))

	if tracer != nil {
		g.Go(func() error {
			return tracer.Shutdown(mainCtx)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("failed to shutdown HTTP server")
	} else {
		logger.Info().Msg("successfully shutdown HTTP server")
	}
}
