FROM golang:1.21.0-alpine3.18 AS build

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev

COPY . .

RUN go build -o ./connectbox-exporter .

FROM alpine:3.18

WORKDIR /app

COPY --from=build /build/connectbox-exporter /app/

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 5000 connectbox-exporter && \
    adduser -S -u 5000 -G connectbox-exporter connectbox-exporter && \
    chown -R connectbox-exporter:connectbox-exporter .

USER connectbox-exporter
EXPOSE 9119

ENTRYPOINT ["/app/connectbox-exporter"]
