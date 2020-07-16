# Start from the latest golang base image
FROM golang:1.13 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -o tyk-sync .


######## Start a new stage from scratch #######
FROM debian:buster-slim

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends \
            curl ca-certificates apt-transport-https debian-archive-keyring gnupg \
 && apt-get purge -y gnupg \
 && apt-get autoremove -y \
 && rm -rf /root/.cache

RUN mkdir -p /opt/tyk-sync
WORKDIR /opt/tyk-sync

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/tyk-sync /opt/tyk-sync/tyk-sync

ENTRYPOINT ["./tyk-sync"]

CMD ["--help"]