# build
FROM golang:1.8-stretch AS build
WORKDIR /go/src/${owner:-github.com/IzakMarais}/reporter
RUN apt-get update && apt-get install -y make git
ADD . .
RUN make build

# create image
FROM debian:stretch
COPY util/texlive.profile /
RUN PACKAGES="wget libswitch-perl" \
    && apt-get update \
    && apt-get install -qq $PACKAGES --no-install-recommends -y \
    && wget -qO- http://mirror.ctan.org/systems/texlive/tlnet/install-tl-unx.tar.gz |tar xz \
    && cd install-tl-* \
    && ./install-tl -profile /texlive.profile \
    # Cleanup
    && rm -rf install-tl-* \
    && apt-get remove --purge -qq $PACKAGES \
    && apt-get autoremove --purge -qq \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /var/tex

COPY --from=build /go/bin/grafana-reporter /usr/local/bin
ENTRYPOINT [ "/usr/local/bin/grafana-reporter" ]
