FROM golang:1.12.1-alpine as builder

WORKDIR /go/src/app
ENV GO111MODULE=on
RUN apk add git
COPY . .
RUN go mod tidy
RUN go build -o kafka-manager-exporter cmd/main.go

FROM alpine:latest

WORKDIR /bin/
COPY --from=builder /go/src/app/kafka-manager-exporter .

ENTRYPOINT ["kafka-manager-exporter"]
