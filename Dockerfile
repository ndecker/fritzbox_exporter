FROM golang:1.8-alpine

ADD . $GOPATH/src/github.com/mxschmitt/fritzbox_exporter

RUN apk add --no-cache git
RUN go get -v github.com/mxschmitt/fritzbox_exporter

EXPOSE 9133

ENTRYPOINT ["fritzbox_exporter"]
CMD [""]

