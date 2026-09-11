# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
      -trimpath \
      -ldflags="-s -w -X main.version=$VERSION -X main.commit=$VCS_REF -X main.buildDate=$BUILD_DATE" \
      -o /out/devops-may-cry ./cmd/devops-may-cry

FROM gcr.io/distroless/static-debian12:nonroot@sha256:1b7b9f0f0e0a1d2155f531db587cc48ec26aaf97ab64364225f5bf18a054e66a

ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="DEVOPS MAY CRY" \
      org.opencontainers.image.description="Production-oriented Kubernetes laboratory API" \
      org.opencontainers.image.source="https://github.com/boost-mentor/kubernetes-production-labs" \
      org.opencontainers.image.version="$VERSION" \
      org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.created="$BUILD_DATE" \
      org.opencontainers.image.licenses="MIT"

WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/devops-may-cry /devops-may-cry

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/devops-may-cry"]
