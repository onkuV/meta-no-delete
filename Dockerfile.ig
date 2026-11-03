FROM golang:1-alpine3.24 AS builder

RUN apk add --no-cache git ca-certificates build-base su-exec olm-dev

COPY . /build
WORKDIR /build
RUN ./build-ig.sh

FROM alpine:3.24

ENV UID=1337 \
    GID=1337

RUN apk add --no-cache ffmpeg su-exec ca-certificates olm bash jq yq-go curl

# docker-run.sh invokes /usr/bin/mautrix-meta, so install the Instagram binary
# under that name (same as Dockerfile.ci does) and keep the real name as an alias.
COPY --from=builder /build/mautrix-instagram /usr/bin/mautrix-meta
RUN ln -s /usr/bin/mautrix-meta /usr/bin/mautrix-instagram
COPY --from=builder /build/docker-run.sh /docker-run.sh
VOLUME /data

CMD ["/docker-run.sh"]
