FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o devcost-api ./cmd/api
EXPOSE 8080
CMD ["./devcost-api"]