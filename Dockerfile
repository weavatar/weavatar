# Build the go application into a binary
FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=1  \
    GOARCH="amd64" \
    GOOS="linux"   \
    GOAMD64="v3"

RUN apk --update add \
    ca-certificates  \
    build-base       \
    pkgconfig        \
    vips-dev         \
    vips-cpp         \
    vips-heif        \
    vips-jxl         \
    vips-magick      \
    vips-poppler

WORKDIR /app
COPY . ./

RUN go mod tidy
RUN go build -ldflags "-s -w" -o app ./cmd/app

# Run the binary on an empty container
FROM alpine

RUN apk --update add \
    ca-certificates  \
    build-base       \
    pkgconfig        \
    vips-dev         \
    vips-cpp         \
    vips-heif        \
    vips-jxl         \
    vips-magick      \
    vips-poppler     \

COPY --from=builder /app/app .
COPY --from=builder /app/config/ ./config/
COPY --from=builder /app/storage/ ./storage/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 3000
ENTRYPOINT ["/app"]
