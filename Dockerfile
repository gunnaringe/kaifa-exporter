FROM golang:1.12-alpine3.9 AS builder
ARG GOOS=linux
ARG GOARCH=arm
ARG GOARM=5

WORKDIR /kaifa
COPY *.go go.mod go.sum /kaifa/

RUN apk add git upx \
 && env CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build -ldflags '-w -s' \
 && upx --best --brute kaifa-exporter

FROM scratch
LABEL maintainer="Gunnar Inge G. Sortland <gunnaringe@gmail.com>"
COPY --from=builder /go/src/github.com/gunnaringe/kaifa-exporter/kaifa-exporter /
EXPOSE 9500
ENTRYPOINT ["/kaifa-exporter"]
