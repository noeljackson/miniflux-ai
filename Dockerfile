FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /miniflux-ai .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && adduser -D -u 1001 appuser
COPY --from=builder /miniflux-ai /usr/local/bin/miniflux-ai
USER appuser
EXPOSE 3000
CMD ["miniflux-ai"]
