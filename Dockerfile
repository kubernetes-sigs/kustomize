FROM golang:1.13-stretch as build
WORKDIR /work
RUN apt-get update
RUN apt-get install gcc curl make -y
RUN curl https://get.helm.sh/helm-v2.15.0-linux-amd64.tar.gz | tar xz
RUN cp linux-amd64/helm /go/bin/
RUN GO111MODULE=on go get github.com/mikefarah/yq@2.4.1
RUN go get github.com/hairyhenderson/gomplate/cmd/gomplate

#has a dependency on make qlik-build-all
COPY bin/linux/kustomize /go/bin/kustomize

FROM debian:stretch
COPY --from=build /go/bin /usr/local/bin
