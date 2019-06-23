# simulation/Dockerfile

# We use the official golang image, which contains all the 
# correct build tools and libraries. Notice `as builder`,
# this gives this container a name that we can reference later on. 
FROM golang:1.12
FROM golang:alpine as builder

RUN apk --no-cache add --update git gcc libc-dev net-tools

# Set our workdir to our current service in the gopath
WORKDIR /app

# Copy the current code into our workdir
RUN git config --global url.https://b88e1d7b0c34e88ad51a6641041bd55890d9f1fc@github.com/.insteadOf https://github.com/
COPY . .
RUN go get github.com/bburch01/FOTAAS/api/...
RUN go get github.com/bburch01/FOTAAS/fotaas-internal/app/simulation/models/...
RUN go get github.com/bburch01/FOTAAS/fotaas-internal/pkg/logging/...
RUN go get github.com/go-sql-driver/mysql/
RUN go get github.com/joho/godotenv/
RUN go get github.com/openzipkin/zipkin-go/
RUN go get github.com/openzipkin/zipkin-go/middleware/grpc/
RUN go get github.com/openzipkin/zipkin-go/reporter/http/
RUN go get google.golang.org/grpc/
RUN go get github.com/google/uuid
RUN go get go.uber.org/zap
RUN go get github.com/jmcvetta/randutil

RUN go mod download

# Build the binary, with a few flags which will allow
# us to run this binary in Alpine. 
RUN GOOS=linux go build -a -o simulation


# Here we're using a second FROM statement, which is strange,
# but this tells Docker to start a new build process with this
# image.
FROM alpine:latest

# Security related package, good to have.
RUN apk --no-cache add ca-certificates procps

# Same as before, create a directory for our app.
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN ls -latr /app


RUN addgroup fotaas && adduser -H -D -G fotaas fotaas
RUN mkdir -p /var/log/fotaas
RUN touch /var/log/fotaas/fotaas.log
RUN chown fotaas /var/log/fotaas/fotaas.log

# Here, instead of copying the binary from our host machine,
# we pull the binary from the container named `builder`, within
# this build context. This reaches into our previous image, finds
# the binary we built, and pulls it into this container. Amazing!
COPY --from=builder /app/simulation .

CMD ./simulation