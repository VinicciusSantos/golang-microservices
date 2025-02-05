# base go image
FROM golang:1.23.5-alpine3.21 AS builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o loggerApp ./cmd/api

RUN chmod +x /app/loggerApp

# build tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/loggerApp /app

CMD [ "/app/loggerApp" ]
