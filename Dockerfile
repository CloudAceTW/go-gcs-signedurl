FROM golang:1.22-bookworm as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY main.go ./
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 go build -v -o go-gcs-signurl

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM gcr.io/distroless/static-debian12
# RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
#     ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

# WORKDIR /app
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/go-gcs-signurl /

# Run the web service on container startup.
CMD ["/go-gcs-signurl"]
