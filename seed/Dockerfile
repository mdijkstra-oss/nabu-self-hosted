# The welcome seed as a container: a Go builder, then a scratch image holding
# only the static binary. No CA bundle — it speaks plain http to storage on the
# compose network and never leaves it.

FROM golang:1.25-alpine AS build

WORKDIR /src

# The manifest is its own layer so that editing a source file does not refetch
# the module cache.
COPY go.mod ./
RUN go mod download

COPY . .

# CGO off is what makes the binary static, which is what allows a scratch final
# stage. -s -w drop the symbol table and DWARF, -trimpath keeps build paths out
# of the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /seed .

FROM scratch

COPY --from=build /seed /seed

# Numeric, because scratch carries no /etc/passwd for a name to resolve against.
USER 65532:65532

ENTRYPOINT ["/seed"]
