# syntax=docker/dockerfile:1.7

FROM golang:1.25.11-bookworm AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=1970-01-01T00:00:00Z
ARG DIRTY=false

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -buildvcs=false \
  -ldflags="-s -w \
    -X github.com/sovereign-l1/l1/cmd/l1d/cmd.appVersion=${VERSION} \
    -X github.com/sovereign-l1/l1/cmd/l1d/cmd.gitCommit=${COMMIT} \
    -X github.com/sovereign-l1/l1/cmd/l1d/cmd.buildDate=${BUILD_DATE} \
    -X github.com/sovereign-l1/l1/cmd/l1d/cmd.dirty=${DIRTY}" \
  -o /out/aetrad ./cmd/l1d

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --create-home --home-dir /home/aetra --shell /usr/sbin/nologin --uid 10001 aetra

COPY --from=builder /out/aetrad /usr/local/bin/aetrad
COPY assets/aetra.png /usr/local/share/aetra.png

ENV DAEMON_HOME=/home/aetra
WORKDIR /home/aetra

EXPOSE 26656 26657 1317 9090

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=6 \
  CMD /usr/local/bin/aetrad status --node tcp://127.0.0.1:26657 --output json >/dev/null 2>&1 || exit 1

USER aetra

ENTRYPOINT ["/usr/local/bin/aetrad"]
CMD ["start"]
