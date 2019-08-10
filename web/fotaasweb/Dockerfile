# fotaasweb/Dockerfile

FROM golang:1.12
FROM golang:alpine as builder

RUN apk --no-cache add --update git gcc libc-dev net-tools

WORKDIR /app/go/src/github.com/bburch01/FOTAAS
COPY . .

ENV GOPATH /app/go
ENV GO111MODULE=auto

RUN go get github.com/gorilla/mux
RUN go get github.com/joho/godotenv/
RUN go get google.golang.org/grpc/
RUN go get github.com/google/uuid
RUN go get go.uber.org/zap

RUN go mod download

WORKDIR /app/go/src/github.com/bburch01/FOTAAS/web/fotaasweb

RUN GOOS=linux go build -i -a -o fotaasweb

FROM alpine:latest

RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache curl

RUN addgroup fotaas && adduser -H -D -G fotaas fotaas
RUN mkdir -p /var/log/fotaas
RUN touch /var/log/fotaas/fotaas.log
RUN chown fotaas /var/log/fotaas/fotaas.log

COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/web/fotaasweb/fotaasweb .
COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/web/fotaasweb/.env .


ENTRYPOINT ./fotaasweb --port 8080
