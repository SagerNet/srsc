FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
LABEL maintainer="nekohasekai <contact-git@sekai.icu>"
COPY . /go/src/github.com/sagernet/srsc
WORKDIR /go/src/github.com/sagernet/srsc
ARG TARGETOS TARGETARCH
ARG GOPROXY=""
ENV GOPROXY ${GOPROXY}
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
RUN set -ex \
    && apk add git build-base \
    && export COMMIT=$(git rev-parse --short HEAD) \
    && export VERSION=$(go run github.com/sagernet/sing-box/cmd/internal/read_tag@latest) \
    && go build -v -trimpath -tags \
        "with_acme" \
        -o /go/bin/srsc \
        -ldflags "-X \"github.com/sagernet/srsc/constant.Version=$VERSION\" -s -w -buildid=" \
        ./cmd/srsc
FROM --platform=$TARGETPLATFORM alpine AS dist
LABEL maintainer="nekohasekai <contact-git@sekai.icu>"
RUN set -ex \
    && apk upgrade \
    && apk add bash tzdata ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/srsc /usr/local/bin/srsc
ENTRYPOINT ["srsc", "-D", "/var/lib/srsc", "-C", "/etc/srsc", "run"]

