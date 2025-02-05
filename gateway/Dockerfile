FROM golang:1.23.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go get ./...
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app /app
EXPOSE 8080:8080
CMD "/app/main"