# syntax=docker/dockerfile:1

#
# Explicitly use the current platform as our build platform, so buildx
# doesn't emulate the target
#
FROM --platform=${BUILDPLATFORM} golang:1.24 AS builder

# Builds deps separately for caching purposes
COPY go.mod .
RUN go mod download

# Build app
# CGO_ENABLED disables c-interop and forces a static link. This
# is great if you can get away with it!
# The linker flags "-s -w" strip symbols and debug to make the image smaller.
# You may or may not want to use these
COPY main.go .
COPY tpl.html .

ARG TARGETPLATFORM
RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build \
    -ldflags="-s -w" -o /app/main

#
# Release stage - scratch gives us an empty docker base so we have nothing
# in it apart from our app!
#
FROM scratch

# Set the workdir and add our app
WORKDIR /app
COPY --from=builder /app/main /app/main

CMD ["/app/main"]
