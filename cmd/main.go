package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/andychao217/callhome"
	"github.com/andychao217/callhome/api"
	"github.com/andychao217/callhome/internal"
	jaegerClient "github.com/andychao217/callhome/internal/clients/jaeger"
	"github.com/andychao217/callhome/internal/clients/postgres"
	"github.com/andychao217/callhome/internal/env"
	"github.com/andychao217/callhome/internal/server"
	httpserver "github.com/andychao217/callhome/internal/server/http"
	"github.com/andychao217/callhome/timescale"
	"github.com/andychao217/callhome/timescale/tracing"
	stracing "github.com/andychao217/callhome/tracing"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "callhome"
	envPrefix      = "MG_CALLHOME_"
	envPrefixHttp  = "MG_CALLHOME_"
	defSvcHttpPort = "8855"
)

type config struct {
	LogLevel       string `env:"MG_CALLHOME_LOG_LEVEL"       envDefault:"info"`
	JaegerURL      string `env:"MG_JAEGER_URL"               envDefault:"http://jaeger:14268/api/traces"`
	IPDatabaseFile string `env:"MG_CALLHOME_IP_DB"           envDefault:"./IP2LOCATION-LITE-DB5.BIN"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := newLogger(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to init logger: %s", err.Error()))
	}
	timescaleDB, err := postgres.Setup(envPrefix, timescale.Migration())
	if err != nil {
		log.Fatalf("failed to setup timescale db : %s", err)
	}

	tp, err := jaegerClient.NewProvider(svcName, cfg.JaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()
	tracer := tp.Tracer(svcName)

	svc, err := newService(ctx, logger, cfg.IPDatabaseFile, timescaleDB, tracer)
	if err != nil {
		log.Fatalf("failed to initialize service: %s", err)
		return
	}

	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		return
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, tp, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(ctx context.Context, logger *slog.Logger, ipDB string, timescaleDB *sqlx.DB, tracer trace.Tracer) (callhome.Service, error) {
	timescaleRepo := timescale.New(timescaleDB)
	timescaleRepo = tracing.New(tracer, timescaleRepo)
	locSvc, err := callhome.NewLocationService(ipDB)
	if err != nil {
		return nil, err
	}
	locSvc = stracing.NewLocationService(tracer, locSvc)
	svc := callhome.New(timescaleRepo, locSvc)
	svc = stracing.NewService(tracer, svc)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	svc = api.LoggingMiddleware(svc, logger)
	return svc, nil
}

// New returns wrapped logger.
func newLogger(w io.Writer, levelText string) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		return &slog.Logger{}, fmt.Errorf(`{"level":"error","message":"%s: %s","ts":"%s"}`, err, levelText, time.RFC3339Nano)
	}

	logHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(logHandler), nil
}
