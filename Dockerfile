#  docker build -t disk:v1 .
#  docker run -it --rm disk:v1 

FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .

RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/example ./example/example.go


FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates
COPY --from=builder /usr/share/zoneinfo/America/New_York /usr/share/zoneinfo/America/New_York
ENV TZ America/New_York
ENV YANDEX_DISK_ACCESS_TOKEN 12345678-your-token-paste-here-87654321

WORKDIR /app
COPY --from=builder /app/example /app/example

CMD ["./example"]