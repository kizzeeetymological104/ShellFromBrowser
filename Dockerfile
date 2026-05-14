# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev) -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)" -o /bin/shellfb ./cmd/shellfb

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache \
    bash \
    openssh-client \
    ca-certificates \
    && adduser -D -h /home/shellfb shellfb

COPY --from=builder /bin/shellfb /usr/local/bin/shellfb
COPY config.example.yaml /etc/shellfb/config.yaml

USER shellfb
WORKDIR /home/shellfb

EXPOSE 8080

ENTRYPOINT ["shellfb"]
CMD ["--config", "/etc/shellfb/config.yaml"]
