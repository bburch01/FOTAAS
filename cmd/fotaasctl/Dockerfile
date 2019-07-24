# fotaasctl/Dockerfile

FROM golang:1.12
FROM golang:alpine as builder

RUN apk --no-cache add --update git gcc libc-dev net-tools

WORKDIR /app/go/src/github.com/bburch01/FOTAAS
COPY . .

ENV GOPATH /app/go
ENV GO111MODULE=auto

RUN go get github.com/spf13/cobra/
RUN go get github.com/mitchellh/go-homedir/
RUN go get github.com/spf13/viper/
RUN go get github.com/joho/godotenv/
RUN go get github.com/google/uuid
RUN go get google.golang.org/grpc/
RUN go get github.com/briandowns/spinner

RUN go mod download

WORKDIR /app/go/src/github.com/bburch01/FOTAAS/cmd/fotaasctl

RUN GOOS=linux go build -a -o fotaasctl

FROM alpine:latest

RUN apk --no-cache add ca-certificates procps

WORKDIR /app/go/src/github.com/bburch01/FOTAAS/cmd/fotaasctl

RUN addgroup fotaas && adduser -H -D -G fotaas fotaas
RUN mkdir -p /var/log/fotaas
RUN touch /var/log/fotaas/fotaas.log
RUN chown fotaas /var/log/fotaas/fotaas.log

COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/cmd/fotaasctl/fotaasctl .
COPY --from=builder /app/go/src/github.com/bburch01/FOTAAS/cmd/fotaasctl/.env .

CMD tail -f /dev/null