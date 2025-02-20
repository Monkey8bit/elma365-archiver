FROM golang:1.23.3-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

CMD ["./main"]