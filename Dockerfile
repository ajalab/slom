ARG GO_VERSION=1.23

FROM golang:${GO_VERSION} AS build

WORKDIR /build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,target=. \
    go build -o /slom

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build /slom /slom

ENTRYPOINT ["/slom"]
