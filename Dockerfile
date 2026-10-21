# syntax=docker/dockerfile:1
# rafaelgaspar/otbr-insight — single static binary on Debian forkly.
FROM docker.io/library/debian:forky-slim@sha256:4cee6ff3ebc49e32d55a857e25164ef8e5f8fecff2a186c6393f7e8bd8c23b53 AS golang-forky
ARG TARGETARCH
ENV DEBIAN_FRONTEND=noninteractive

RUN sed -i 's|http://|https://|g' /etc/apt/sources.list /etc/apt/sources.list.d/*.sources 2>/dev/null; \
    apt-get update -o Acquire::https::Verify-Peer=false \
      && apt-get install -y --no-install-recommends -o Acquire::https::Verify-Peer=false ca-certificates curl git \
      && rm -rf /var/lib/apt/lists/*

RUN apt-get update \
  && apt-get upgrade -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
  && apt-get install -y --no-install-recommends ca-certificates curl git \
  && rm -rf /var/lib/apt/lists/* \
  && GO_VERSION="$(curl -fsSL https://go.dev/VERSION?mode=text | head -1)" \
  && curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-${TARGETARCH}.tar.gz" | tar -C /usr/local -xzf - \
  && /usr/local/go/bin/go version

ENV PATH=/usr/local/go/bin:$PATH

FROM golang-forky AS build
ARG TARGETARCH

WORKDIR /src
COPY . .

ENV CGO_ENABLED=0 GOTOOLCHAIN=auto
RUN GOARCH="${TARGETARCH}" go build -trimpath -ldflags="-s -w" -o /out/otbr-insight ./cmd/otbr-insight

FROM docker.io/library/debian:forky-slim@sha256:4cee6ff3ebc49e32d55a857e25164ef8e5f8fecff2a186c6393f7e8bd8c23b53

RUN sed -i 's|http://|https://|g' /etc/apt/sources.list /etc/apt/sources.list.d/*.sources 2>/dev/null; \
    apt-get update -o Acquire::https::Verify-Peer=false \
      && apt-get install -y --no-install-recommends -o Acquire::https::Verify-Peer=false ca-certificates curl \
      && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/otbr-insight /usr/local/bin/otbr-insight

# Daemon socket is root:root mode 0755; connect() needs write.
USER 0

EXPOSE 8088

ENTRYPOINT ["/usr/local/bin/otbr-insight"]
