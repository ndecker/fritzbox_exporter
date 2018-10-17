FROM golang AS build-env

LABEL maintainer="Max Schmitt <max@schmitt.mx>"
LABEL description="FRITZ!Box Prometheus exporter"

ADD . /go/src/github.com/mxschmitt/fritzbox_exporter

RUN cd /go/src/github.com/mxschmitt/fritzbox_exporter/cmd/exporter && \
    go get ./... && \
    CGO_ENABLED=0 go build -o /exporter

FROM alpine

RUN apk update && apk add ca-certificates

COPY --from=build-env /exporter /

EXPOSE 9133

ENTRYPOINT ["/exporter"]
