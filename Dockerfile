FROM golang:1.15-alpine

ADD . $GOPATH/src/github.com/ndecker/fritzbox_exporter

RUN apk add --no-cache git
RUN go get -v github.com/ndecker/fritzbox_exporter

EXPOSE 9133

ENTRYPOINT ["fritzbox_exporter"]
CMD [""]

