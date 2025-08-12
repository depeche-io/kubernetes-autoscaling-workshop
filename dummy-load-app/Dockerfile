# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22

############################
# Build stage
############################
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

# Enable module mode and faster, repeatable builds
ENV CGO_ENABLED=0
RUN apk add --no-cache git

# Copy mod files first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build a static binary for the target platform
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -tags netgo \
      -ldflags='-s -w -extldflags "-static"' \
      -o /out/server .

############################
# Run stage
############################
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]

# Example default flags (optional). Override with `docker run ... -- -cpu=50 ...`
# CMD ["-cpu=50", "-mem=64", "-time=100", "-jitter=0.2"]

