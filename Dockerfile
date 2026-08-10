# docker build -t disk:demo .
# docker run -it --rm -e YANDEX_DISK_ACCESS_TOKEN=<token> disk:demo

FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk add --no-cache tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/demo ./examples/demo


FROM alpine

RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /usr/share/zoneinfo/Europe/Moscow
ENV TZ=Europe/Moscow

WORKDIR /app
COPY --from=builder /app/demo /app/demo

CMD ["./demo"]
