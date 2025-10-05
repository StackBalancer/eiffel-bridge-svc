FROM docker.io/golang:1.23-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o adapter ./cmd/adapter

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/adapter /adapter
EXPOSE 8080
ENTRYPOINT ["/adapter"]
