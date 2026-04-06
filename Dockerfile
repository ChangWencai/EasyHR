FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /easyhr ./cmd/server/

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /easyhr .
COPY config/config.yaml ./config/config.yaml
EXPOSE 8080
ENTRYPOINT ["/easyhr"]
