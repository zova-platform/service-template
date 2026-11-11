# service-template

Template for ZOVA Go services. It comes with config, JSON logging, OpenTelemetry, health endpoints, a gRPC server with interceptors and graceful shutdown already wired from [platform-go](https://github.com/zova-platform/platform-go).

## Create a service from this template

Delete this section from the new repo's README when done.

1. Click **Use this template** on GitHub and create the repo in the `zova-platform` org, named after the service, for example `trip-service`.
2. Clone it and run `task rename NEW=trip-service`. This renames the Go module, the `cmd/` directory and every mention of `service-template`.
3. Check with `task check`, commit, and push.
4. Fill in the sections below for the new service.

## Run

```sh
task run
```

Then:

| What | Where |
| --- | --- |
| gRPC | `localhost:9090` (standard `grpc.health.v1.Health`) |
| Liveness | `http://localhost:8081/healthz` |
| Readiness | `http://localhost:8081/readyz` |
| Metrics | `http://localhost:8081/metrics` |

To send traces and metrics to the local stack from the `deploy` repo, run `task up` there and start the service with `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317`.

Run in Docker:

```sh
docker build -t service-template .
docker run --rm -p 9090:9090 -p 8081:8081 service-template
```

The image is distroless and runs as a non-root user. Logs report version `dev` unless the image is built with `--build-arg VERSION=<version>`.

## Environment variables

| Variable | Default | Meaning |
| --- | --- | --- |
| `SERVICE_NAME` | `service-template` | Name used in logs and telemetry |
| `GRPC_ADDR` | `:9090` | gRPC listen address |
| `HTTP_ADDR` | `:8081` | Address for `/healthz`, `/readyz` and `/metrics` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `SHUTDOWN_TIMEOUT` | `30s` | How long in-flight calls get to finish on shutdown |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | OpenTelemetry Collector, read by the OpenTelemetry SDK. In compose use `http://otel-collector:4317` |

A service that uses a database, Redis or NATS adds `DATABASE_URL`, `REDIS_URL` and `NATS_URL`. Calls to other services use `<NAME>_SERVICE_ADDR`, for example `LOCATION_SERVICE_ADDR=location-service:9090`.

## gRPC methods

None yet. List every method of the service here.

## Events

| Direction | Subject | Payload |
| --- | --- | --- |
| Publishes | none | |
| Consumes | none | |

## Tables owned

None yet. List the tables this service owns here. No other service may read or write them.

## Layout

```
cmd/service-template/   main: reads config and starts everything
internal/config/        environment variables
internal/domain/        entities and business rules, no infrastructure imports
internal/app/           use cases
internal/grpc/          gRPC handlers that translate proto to domain
internal/events/        NATS publishers and consumers
internal/store/         Postgres and Redis
migrations/             SQL migrations
queries/                SQL for sqlc
```

## Tasks

| Task | What it does |
| --- | --- |
| `task run` | Run the service |
| `task build` | Build `bin/service-template` |
| `task test` | Tests with the race detector |
| `task lint` | golangci-lint |
| `task check` | Build, lint and test |
| `task tidy` | `go mod tidy` |
| `task rename NEW=<name>` | Rename the template into a real service |

## CI

`.github/workflows/ci.yml` calls the shared workflows in `ci-workflows@v1`: lint, tests with `-race` and govulncheck on every push and pull request. On pushes to `main` and version tags it also builds the image, scans it with Trivy and pushes it to `ghcr.io/zova-platform/service-template`.
