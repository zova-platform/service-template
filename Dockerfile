FROM golang:1.26 AS build
ARG SERVICE=service-template
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/service /service
EXPOSE 9090 8081
USER nonroot:nonroot
ENTRYPOINT ["/service"]
