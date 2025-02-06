# base go image
FROM golang:1.23.5-alpine3.21 AS builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o mailApp ./cmd/api

RUN chmod +x /app/mailApp

# build tiny docker image
FROM alpine:latest

RUN mkdir /templates

COPY --from=builder /app/mailApp .
COPY --from=builder /app/templates /templates

CMD [ "/mailApp" ]
