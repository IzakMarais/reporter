# build
FROM golang:alpine3.13 AS build
WORKDIR /go/src/${owner:-github.com/IzakMarais}/reporter
RUN apk update && apk add make git
ADD . .
RUN make build

# create image
FROM alpine:3.13.3
COPY util/texlive.profile /

RUN PACKAGES="wget" \
        && apk update \
        && apk add $PACKAGES \
        && apk add ca-certificates perl freetype fontconfig \
        && wget -qO- \
          "https://github.com/yihui/tinytex/raw/master/tools/install-unx.sh" | \
          sh -s - --admin --no-path \
        && mv ~/.TinyTeX /opt/TinyTeX \
        && /opt/TinyTeX/bin/*/tlmgr path add \
        && tlmgr path add \
        && chown -R root:root /opt/TinyTeX \
        && chmod -R g+w /opt/TinyTeX \
        && chmod -R g+wx /opt/TinyTeX/bin \
        && tlmgr install epstopdf-pkg \
        && apk del $PACKAGES \
        && rm -rf /var/cache/apk/*

COPY --from=build /go/bin/grafana-reporter /usr/local/bin
ENTRYPOINT [ "/usr/local/bin/grafana-reporter" ]
