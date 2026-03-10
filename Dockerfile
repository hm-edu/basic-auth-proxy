# syntax=docker/dockerfile:1

FROM golang:1.26 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/basic-auth-proxy .

FROM scratch
WORKDIR /app

COPY --from=builder /out/basic-auth-proxy /usr/local/bin/basic-auth-proxy
COPY config.example.yaml /app/config.yaml

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/basic-auth-proxy"]
CMD ["-config", "/app/config.yaml"]
