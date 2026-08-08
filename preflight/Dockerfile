# Preflight as a container: a Go builder, a stage that captures the service
# table out of the stack's own dragoman image, then a scratch image holding the
# static binary and that captured table.

FROM golang:1.25-alpine AS build

WORKDIR /src

# The manifests are their own layer so that editing a source file does not
# refetch the module cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /preflight .

# `dragoman` is an additional build context compose points at the dragoman
# service's image, so the table captured here is the embedded table of the
# exact image the stack runs — the two can only skew on a partial rebuild.
FROM alpine AS capture

COPY --from=dragoman /dragoman /dragoman
RUN /dragoman config > /dragoman.yaml

FROM scratch

COPY --from=build /preflight /preflight
COPY --from=capture /dragoman.yaml /etc/preflight/dragoman.yaml

ENTRYPOINT ["/preflight"]
