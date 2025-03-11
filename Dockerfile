# Build the go application into a binary
FROM golang:alpine AS builder

ENV GO111MODULE=on                                                                              \
    GOARCH="amd64"                                                                              \
    GOOS="linux"                                                                                \
    GOAMD64="v3"                                                                                \
    CGO_ENABLED=1                                                                               \
    CGO_CFLAGS="-fno-builtin-malloc -fno-builtin-calloc -fno-builtin-realloc -fno-builtin-free" \
    CGO_LDFLAGS="-ljemalloc"

RUN apk --update add \
    ca-certificates  \
    build-base       \
    pkgconfig        \
    jemalloc-dev     \
    upx              \
    vips-dev         \
    vips-heif        \
    vips-jxl         \
    vips-magick      \
    vips-poppler

WORKDIR /app
COPY . ./

RUN go mod tidy
RUN go build -ldflags "-s -w" -o app ./cmd/app
RUN upx --best --lzma app

# Run the binary on an empty container
FROM alpine

RUN apk --update add \
    ca-certificates  \
    jemalloc         \
    vips             \
    vips-heif        \
    vips-jxl         \
    vips-magick      \
    vips-poppler

COPY --from=builder /app/app .
COPY --from=builder /app/config/ ./config/
COPY --from=builder /app/storage/ ./storage/

EXPOSE 3000
ENTRYPOINT ["/app"]
