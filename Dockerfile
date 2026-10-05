# Build the go application into a binary
FROM golang:alpine AS builder

ENV GO111MODULE=on \
    GOARCH="amd64" \
    GOOS="linux"   \
    GOAMD64="v3"   \
    CGO_ENABLED=0

WORKDIR /app
COPY . ./

RUN go mod tidy
RUN go build -ldflags "-s -w" -o app ./cmd/app

# Run the binary on an empty container
FROM alpine

RUN apk --update add ca-certificates

COPY --from=builder /app/app .
COPY --from=builder /app/config/ ./config/
COPY --from=builder /app/storage/ ./storage/

EXPOSE 3000
ENTRYPOINT ["/app"]
