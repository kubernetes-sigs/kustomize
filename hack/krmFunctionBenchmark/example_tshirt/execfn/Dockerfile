FROM alpine:latest

ENV BUILD_HOME=/usr/local/build
RUN apk update && apk add --no-cache git go

RUN mkdir -p $BUILD_HOME

WORKDIR $BUILD_HOME

RUN git clone https://github.com/kubernetes-sigs/kustomize.git .
RUN git checkout tags/kustomize/v3.6.1
WORKDIR $BUILD_HOME/functions/examples/injection-tshirt-sizes/image/

ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o tshirt .
