FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.work ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service ./cmd/auth-service

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/auth-service /app/auth-service
EXPOSE 8081
ENTRYPOINT ["/app/auth-service"]
