FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

# Deploy the application binary into a lean image
FROM alpine:latest

WORKDIR /

COPY --from=build /app/main /app/main
COPY --from=build /app/configs/prod.yaml /app/configs/prod.yaml

ENTRYPOINT ["/app/main", "--config=/app/configs/prod.yaml"]