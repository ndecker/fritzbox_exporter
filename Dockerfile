FROM golang:1.19.4-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=0

# cache dependencies
# from https://github.com/montanaflynn/golang-docker-cache
COPY go.mod go.sum ./
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs -r go get

COPY . ./

RUN go build -v .

FROM busybox

COPY --from=builder /build/fritzbox_exporter /app/fritzbox_exporter

EXPOSE 9133
ENTRYPOINT /app/fritzbox_exporter
