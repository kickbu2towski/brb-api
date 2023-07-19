FROM golang:1.19-alpine

WORKDIR /app

COPY . .

RUN go build -o app ./cmd/api

EXPOSE 6969

CMD ["./app"]
