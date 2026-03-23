FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o udp-client ./cmd/client

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/udp-client /udp-client
USER 65534:65534
ENTRYPOINT ["/udp-client"]
CMD ["-signal-address=udp-server:3986", "-local-address=0.0.0.0:0", "-room-key=default"]
