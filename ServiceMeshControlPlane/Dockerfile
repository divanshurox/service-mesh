FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . ./
RUN go mod download
RUN go build -o /app/main ./cmd/servicemeshcontrolplane

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main /app/main
EXPOSE 8090
CMD ["/app/main"]