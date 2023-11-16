# build

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

# main image

FROM --platform=$TARGETPLATFORM alpine:3.18.4

RUN apk add --update --no-cache tini

COPY --from=build /src/netexp /usr/local/bin/netexp

ENTRYPOINT [ "tini", "--", "netexp" ]
