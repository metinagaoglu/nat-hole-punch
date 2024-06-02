# Go builder
FROM golang:1.22.3 as builder
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 3986
RUN env CGO_ENABLED=0 go build -o udp-server cmd/main.go 

# Final image
FROM scratch
WORKDIR /app
COPY --from=builder /app/udp-server /udp-server
EXPOSE 3986

ENTRYPOINT ["/udp-server"]