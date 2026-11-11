package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	appconfig "github.com/zova-platform/service-template/internal/config"
)

func unreachableCollector(t *testing.T) {
	t.Helper()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("OTEL_EXPORTER_OTLP_TIMEOUT", "500")
}

func freeAddr(t *testing.T) string {
	t.Helper()
	var lc net.ListenConfig
	lis, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := lis.Addr().String()
	if err := lis.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

func get(t *testing.T, url string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func waitFor(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if get(t, url) == http.StatusOK {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s did not become ready", url)
}

func TestServeExposesProbesAndGRPCHealthThenStopsCleanly(t *testing.T) {
	unreachableCollector(t)

	cfg := appconfig.Config{
		ServiceName:     "test-service",
		GRPCAddr:        freeAddr(t),
		HTTPAddr:        freeAddr(t),
		LogLevel:        "error",
		ShutdownTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, cfg) }()

	base := "http://" + cfg.HTTPAddr
	waitFor(t, base+"/healthz")
	waitFor(t, base+"/readyz")
	if code := get(t, base+"/metrics"); code != http.StatusOK {
		t.Fatalf("/metrics = %d, want 200", code)
	}

	conn, err := grpc.NewClient(cfg.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	callCtx, callCancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer callCancel()
	resp, err := healthpb.NewHealthClient(conn).Check(callCtx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("grpc health = %v, want SERVING", resp.GetStatus())
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve returned %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("serve did not stop after cancel")
	}
}

func TestServeRejectsInvalidLogLevel(t *testing.T) {
	cfg := appconfig.Config{LogLevel: "loud"}
	if err := serve(t.Context(), cfg); err == nil {
		t.Fatal("want an error for an invalid LOG_LEVEL")
	}
}

func TestServeFailsWhenPortIsTaken(t *testing.T) {
	unreachableCollector(t)
	var lc net.ListenConfig
	taken, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = taken.Close() }()

	cfg := appconfig.Config{
		ServiceName:     "test-service",
		GRPCAddr:        taken.Addr().String(),
		HTTPAddr:        "127.0.0.1:0",
		LogLevel:        "error",
		ShutdownTimeout: time.Second,
	}
	if err := serve(t.Context(), cfg); err == nil {
		t.Fatal("want an error when the gRPC port is already in use")
	}
}
