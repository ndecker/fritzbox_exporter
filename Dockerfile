FROM golang:1.15-alpine AS builder

WORKDIR /build
COPY . /build/

ENV CGO_ENABLED=0

RUN go build -v .

FROM busybox

COPY --from=builder /build/fritzbox_exporter /usr/local/bin/fritzbox_exporter

EXPOSE 9133
ENTRYPOINT /usr/local/bin/fritzbox_exporter
