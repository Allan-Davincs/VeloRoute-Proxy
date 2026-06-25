FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /veloroute ./cmd/veloroute

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /veloroute .
COPY config.yaml .
EXPOSE 8080 9090 9091
ENTRYPOINT ["./veloroute", "--config", "config.yaml"]
