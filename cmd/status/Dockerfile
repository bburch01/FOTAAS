# telemetry/Dockerfile

FROM golang:1.12
FROM golang:alpine as builder

RUN apk --no-cache add --update git gcc libc-dev net-tools

WORKDIR /app/go/src/github.com/bburch01/FOTAAS
COPY . .

ENV GOPATH /app/go
ENV GO111MODULE=auto

RUN go get github.com/go-sql-driver/mysql/
RUN go get github.com/joho/godotenv/
RUN go get github.com/openzipkin/zipkin-go/
RUN go get github.com/openzipkin/zipkin-go/middleware/grpc/
RUN go get github.com/openzipkin/zipkin-go/reporter/http/
RUN go get google.golang.org/grpc/
RUN go get github.com/google/uuid
RUN go get go.uber.org/zap
RUN go mod download

WORKDIR /app/go/src/github.com/bburch01/FOTAAS/cmd/status

RUN GOOS=linux go build -a -o status

FROM alpine:latest

RUN apk --no-cache add ca-certificates procps

#WORKDIR /app/go/src/github.com/bburch01/FOTAAS/cmd/telemetry

RUN addgroup fotaas && adduser -H -D -G fotaas fotaas
RUN mkdir -p /var/log/fotaas
RUN touch /var/log/fotaas/fotaas.log
RUN chown fotaas /var/log/fotaas/fotaas.log

COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/cmd/status/status .
COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/cmd/status/.env .

CMD ./status