package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/zova-platform/platform-go/config"
	"github.com/zova-platform/platform-go/grpcx"
	httphealth "github.com/zova-platform/platform-go/health"
	"github.com/zova-platform/platform-go/logx"
	"github.com/zova-platform/platform-go/otelx"
	"github.com/zova-platform/platform-go/run"

	appconfig "github.com/zova-platform/service-template/internal/config"
)

var version = "dev"

func main() {
	cfg, err := config.Load[appconfig.Config]()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := serve(context.Background(), cfg); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, cfg appconfig.Config) error {
	level, err := logx.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("LOG_LEVEL: %w", err)
	}
	logger := logx.New(logx.Options{Level: level}).With("service", cfg.ServiceName, "version", version)
	slog.SetDefault(logger)

	shutdownTelemetry, err := otelx.Setup(ctx, otelx.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}
	handedOff := false
	defer func() {
		if !handedOff {
			_ = shutdownTelemetry(context.WithoutCancel(ctx))
		}
	}()

	probes := httphealth.New(httphealth.Options{})

	grpcServer := grpcx.NewServer(grpcx.ServerOptions{Logger: logger})
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())

	var lc net.ListenConfig
	grpcLis, err := lc.Listen(ctx, "tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen gRPC: %w", err)
	}
	httpLis, err := lc.Listen(ctx, "tcp", cfg.HTTPAddr)
	if err != nil {
		_ = grpcLis.Close()
		return fmt.Errorf("listen HTTP: %w", err)
	}

	runner := run.New(run.Options{
		Logger:          logger,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Health:          probes,
	})
	handedOff = true
	runner.OnShutdown("telemetry", func(ctx context.Context) error {
		if err := shutdownTelemetry(ctx); err != nil {
			logger.Warn("telemetry flush failed", "error", err)
		}
		return nil
	})
	runner.AddHTTP("http", &http.Server{
		Handler:           probes.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}, httpLis)
	runner.AddGRPC("grpc", grpcServer, grpcLis)

	logger.Info("service starting", "grpc", grpcLis.Addr().String(), "http", httpLis.Addr().String())
	return runner.Run(ctx)
}
