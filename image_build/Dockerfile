FROM ubuntu:18.04

ENV NAME=terraform
ENV VERSION=0.11.13
ENV HASHICORP_RELEASES=https://releases.hashicorp.com/terraform

RUN cd /tmp &&\
    apt-get update &&\ 
    apt-get install -y wget zip iputils-ping net-tools git &&\
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_linux_amd64.zip && \
    wget ${HASHICORP_RELEASES}/${VERSION}/${NAME}_${VERSION}_SHA256SUMS && \
    unzip -d /bin ${NAME}_${VERSION}_linux_amd64.zip && \
    rm -rf /tmp/build
