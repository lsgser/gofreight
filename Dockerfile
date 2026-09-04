FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /gofreight ./cmd/gofreight

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /gofreight /usr/local/bin/gofreight
EXPOSE 3000
ENV GOFREIGHT_ENV=production
CMD ["gofreight"]
