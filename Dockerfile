# One Dockerfile for all Go services. Pick the binary with a build arg:
#   docker build --build-arg SERVICE=ingestion_api .
#   docker build --build-arg SERVICE=event-consumer .
#   docker build --build-arg SERVICE=dlq-recorder .

# ---- build stage: compile a static binary ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# Download modules first so this layer is cached when only code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG SERVICE=ingestion_api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

# ---- runtime stage: small image, non-root user ----
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/app /usr/local/bin/app
USER app
ENTRYPOINT ["/usr/local/bin/app"]
