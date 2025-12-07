ARG TARGETARCH
ARG TARGETOS
ARG GO_VERSION=1.24

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder

ARG MODULE_VERSION=main

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /slom

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=builder /slom /slom

ENTRYPOINT ["/slom"]
