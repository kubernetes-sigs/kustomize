# build
FROM golang:alpine as builder
ARG VERSION
ARG COMMIT
ARG DATE
RUN mkdir /build 
ADD . /build/
WORKDIR /build/kustomize
RUN CGO_ENABLED=0 GO111MODULE=on go build \
    -ldflags="-s -X sigs.k8s.io/kustomize/api/provenance.version=${VERSION} \
    -X sigs.k8s.io/kustomize/api/provenance.gitCommit=${COMMIT} \
    -X sigs.k8s.io/kustomize/api/provenance.buildDate=${DATE}"

# only copy binary
FROM alpine
COPY --from=builder /build/kustomize /app/
WORKDIR /app
ENTRYPOINT ["./kustomize"]
