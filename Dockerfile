# syntax=docker/dockerfile:1
# rafaelgaspar/otbr-insight — single static binary on Debian forkly.
FROM docker.io/library/golang:bookworm@sha256:32fa283222e2905d9fee971f60c842e5c815fdc036cf7ea599bf36bd7ca8f67f AS build
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
