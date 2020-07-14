FROM alpine:latest

ENV BUILD_HOME=/usr/local/build
RUN apk update && apk add --no-cache git nodejs npm
RUN npm install -g typescript

RUN mkdir -p $BUILD_HOME

WORKDIR $BUILD_HOME

RUN git clone https://github.com/GoogleContainerTools/kpt-functions-sdk.git .
RUN git checkout tags/release-kpt-functions-v0.14.2
WORKDIR $BUILD_HOME/ts/hello-world/
RUN npm install
RUN npm run build
