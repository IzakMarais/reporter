# build
FROM golang:alpine AS build
WORKDIR /go/src/${owner:-github.com/IzakMarais}/reporter
ADD . .
RUN apk update && apk add make git && make build FORK=${owner:-github.com/IzakMarais}

# create image
FROM alpine:latest
COPY --from=build /go/bin/grafana-reporter /usr/local/bin
ENTRYPOINT [ "/usr/local/bin/grafana-reporter" ]
