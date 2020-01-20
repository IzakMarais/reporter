# build
FROM golang:1.8-stretch AS build
WORKDIR /go/src/${owner:-github.com/IzakMarais}/reporter
RUN apt-get update && apt-get -y install make git
ADD . .
RUN make build

# create image
FROM debian:stretch
COPY util/texlive.profile /

RUN PACKAGES="wget libswitch-perl" \
        && apt-get update \
        && apt-get install -y -qq $PACKAGES --no-install-recommends \
        && apt-get install -y ca-certificates --no-install-recommends \
        && wget -qO- \
          "https://github.com/yihui/tinytex/raw/master/tools/install-unx.sh" | \
          sh -s - --admin --no-path \
        && mv ~/.TinyTeX /opt/TinyTeX \
        && /opt/TinyTeX/bin/*/tlmgr path add \
        && tlmgr path add \
        && chown -R root:staff /opt/TinyTeX \
        && chmod -R g+w /opt/TinyTeX \
        && chmod -R g+wx /opt/TinyTeX/bin \
        && tlmgr install epstopdf-pkg \
        # Cleanup
        && apt-get remove --purge -qq $PACKAGES \
        && apt-get autoremove --purge -qq \
        && rm -rf /var/lib/apt/lists/*


COPY --from=build /go/bin/grafana-reporter /usr/local/bin
ENTRYPOINT [ "/usr/local/bin/grafana-reporter" ]
