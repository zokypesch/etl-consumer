FROM golang:1.13.3

# DO_NOT_REPLACE

# Set Arguments
ARG APP_NAME=platform

# Set go bin which doesn't appear to be set already.
ENV GOBIN /go/bin

# build directories
ENV SRC_DIR="/go/src/github.com/zokypesch/etl-consumer"
RUN mkdir /app
RUN mkdir -p $SRC_DIR

# Copy current directory
COPY ./go.mod ./go.sum $SRC_DIR/
WORKDIR $SRC_DIR

# Build my app
COPY . $SRC_DIR/

# go mod download
RUN go mod download

RUN go build -o /app/main .
CMD ["/app/main"]

