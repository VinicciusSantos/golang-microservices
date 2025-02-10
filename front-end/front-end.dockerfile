FROM alpine:latest

RUN mkdir -p /app

COPY frontApp /app

CMD ["/app/frontApp"]
