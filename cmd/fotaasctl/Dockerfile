# fotaasctl/Dockerfile

# We use the official golang image, which contains all the 
# correct build tools and libraries. Notice `as builder`,
# this gives this container a name that we can reference later on. 
FROM golang:1.12
FROM golang:alpine as builder

RUN apk --no-cache add --update git gcc libc-dev net-tools

# Set our workdir to our current service in the gopath
WORKDIR /app

# Copy the current code into our workdir
RUN git config --global url.https://bba4029d32b08e2cac769354b2663530b612c144@github.com/.insteadOf https://github.com/
COPY . .
RUN pwd;ls -a

RUN go get github.com/bburch01/FOTAAS/api/...
RUN go get github.com/spf13/cobra/
RUN go get github.com/mitchellh/go-homedir/
RUN go get github.com/spf13/viper/
RUN go get github.com/joho/godotenv/
RUN go get google.golang.org/grpc/
RUN go mod download

# Build the binary, with a few flags which will allow
# us to run this binary in Alpine. 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fotaasctl

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
RUN pwd;ls -a

RUN addgroup fotaas && adduser -H -D -G fotaas fotaas
RUN mkdir -p /var/log/fotaas
RUN touch /var/log/fotaas/fotaas.log
RUN chown fotaas /var/log/fotaas/fotaas.log

# Here, instead of copying the binary from our host machine,
# we pull the binary from the container named `builder`, within
# this build context. This reaches into our previous image, finds
# the binary we built, and pulls it into this container. Amazing!
COPY --from=builder /app/fotaasctl .
RUN mkdir -p /src/github.com/bburch01/FOTAAS/
COPY --from=builder /app/.env /src/github.com/bburch01/FOTAAS/
RUN ls -la /src/github.com/bburch01/FOTAAS/

# Need to wait for the database container to start MySQL and create the database if they don't exist.
# Need to wait for the database container to start MySQL and create the database if they don't exist.
# This could take 30 secs!
#
## THE LIFE SAVER
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

## Launch the wait tool and then your application
#CMD ["/wait && ./fotaasctl && tail -f /dev/null"]


# Run the binary as per usual! This time with a binary build in a
# separate container, with all of the correct dependencies and
# run time libraries.
# Using this command to keep the container running
CMD tail -f /dev/null