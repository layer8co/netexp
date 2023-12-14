# ===============
# = build image
# ===============

FROM --platform=$BUILDPLATFORM golang:alpine AS build

WORKDIR /src
COPY . .

ARG TARGETOS
ARG TARGETARCH

ARG GOOS=$TARGETOS
ARG GOARCH=$TARGETARCH
ARG CGO_ENABLED=0

RUN go test -v ./...
RUN go build -trimpath -ldflags '-s -w -buildid='

# ===============
# = main image
# ===============

FROM --platform=$TARGETPLATFORM alpine:3.19

RUN apk add --update --no-cache tini
COPY --from=build /src/netexp /usr/local/bin/netexp
ENTRYPOINT [ "tini", "--", "netexp" ]

COPY --chmod=755 <<-'EOF' /healthcheck.sh
	#!/bin/sh -eu
	listen=${NETEXP_LISTEN:-:9298}
	host=${listen%:*}
	port=${listen#$host}
	content=$(wget --quiet --tries=1 --output-document=- "http://127.0.0.1$port")
	test -n "$(printf '%s\n' "$content" | wc -l)"
EOF
HEALTHCHECK CMD [ "/healthcheck.sh" ]
