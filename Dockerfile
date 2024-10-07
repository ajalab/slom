ARG TARGETARCH
ARG TARGETOS
ARG GO_VERSION=1.23

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder

ARG MODULE_VERSION=main

# Use `go install` to download and build slom rather than run `go build` on the local filesystem,
# so that the executable must contain version build info for `slom version` command.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go install -trimpath github.com/ajalab/slom@$MODULE_VERSION

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /go/bin/slom /slom

ENTRYPOINT ["/slom"]
