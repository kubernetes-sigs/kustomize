FROM docker:19.03.5 AS docker
FROM golang:1.13.4-stretch as build
WORKDIR /work
RUN echo "deb http://deb.debian.org/debian stretch-backports main" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get install gcc curl make -y
RUN apt-get install libgpgme11-dev libassuan-dev libbtrfs-dev libdevmapper-dev -y
RUN curl https://get.helm.sh/helm-v2.15.0-linux-amd64.tar.gz | tar xz
RUN cp linux-amd64/helm /go/bin/
RUN GO111MODULE=on go get github.com/mikefarah/yq/v2
RUN go get github.com/hairyhenderson/gomplate/cmd/gomplate

# this has a dependency on running "make qlik-build-all" before "docker build"
COPY bin/linux/kustomize /go/bin/kustomize

# install troubleshoot for preflight checks
RUN git clone https://github.com/replicatedhq/troubleshoot.git &&\
    cd troubleshoot && make preflight && ls -ltr bin && mv bin/preflight /go/bin &&\
    make support-bundle && ls -ltr bin && mv bin/support-bundle /go/bin &&\
    rm -rf troubleshoot

FROM debian:stretch
RUN apt-get update && apt-get install jq curl git -y && rm -rf /var/lib/apt/lists/*

ENV JFROG_CLI_OFFER_CONFIG=false
RUN curl -fL https://getcli.jfrog.io | sh &&\
    mv jfrog /bin
COPY --from=build /go/bin /usr/local/bin
COPY --from=docker /usr/local/bin/docker /bin/docker

# install porter
ENV PORTER_HOME=/root/.porter
RUN curl https://cdn.deislabs.io/porter/latest/install-linux.sh | bash
RUN echo "export PATH=$PATH:$PORTER_HOME" >> /root/.bashrc
# install porter-mixins
RUN $PORTER_HOME/porter mixin install kustomize -v 0.2-beta-3-0e19ca4 --url https://github.com/donmstewart/porter-kustomize/releases/download
RUN $PORTER_HOME/porter mixin install qliksense -v v0.9.0 --url https://github.com/qlik-oss/porter-qliksense/releases/download