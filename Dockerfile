# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY util/ util/
COPY terranova/ terranova/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM ubuntu:18.04
WORKDIR /
COPY --from=builder /workspace/manager .
#USER nonroot:nonroot

# NAME and VERSION are the name of the software in releases.hashicorp.com
# and the version to download. Example: NAME=terraform VERSION=1.2.3.
#ARG NAME
#ARG VERSION

# Set ARGs as ENV so that they can be used in ENTRYPOINT/CMD
ENV NAME=terraform
#ENV VERSION=0.13.5
ENV VERSION=0.11.13
# This is the location of the releases.
ENV HASHICORP_RELEASES=https://releases.hashicorp.com/terraform
# Create a non-root user to run the software.
#RUN addgroup ${NAME} && \
#    adduser -S -G ${NAME} ${NAME}

# Set up certificates, base tools, and software.
RUN cd /tmp && \
    apt-get update && \
    apt-get install wget zip iputils-ping net-tools -y && \
    #apk add --no-cache ca-certificates curl gnupg libcap openssl su-exec iputils && \
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_linux_amd64.zip && \
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_SHA256SUMS && \
    unzip -d /bin ${NAME}_${VERSION}_linux_amd64.zip && \
    #cd /tmp && \
    rm -rf /tmp/build && \
    mkdir /terraform
    #apk del gnupg openssl && \
    #rm -rf /root/.gnupg

#ADD terraform_template/ /terraform

ENTRYPOINT ["/manager"]
