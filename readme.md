# Grafana reporter

A simple http service that generates *.PDF reports from [Grafana](http://grafana.org/) dashboards.

![demo](demo/report.gif)

## Requirements

Runtime requirements
* `pdflatex` installed and available in PATH.
* a running Grafana instance that it can connect to

Build requirements:
* [golang](https://golang.org/)

## Getting started

### Build and run

Get the source files and dependencies:

    go get github.com/izakmarais/reporter/... github.com/smartystreets/goconvey

Build and install:

    go install -v github.com/izakmarais/reporter/cmd/grafana-reporter

Running without any flags assumes Grafana is reachable at _localhost:3000_:

    ./bin/grafana-reporter

Query available flags:

    ./bin/grafana-reporter --help

### Test

The unit tests can be run using the go tool:

    go test -v github.com/izakmarais/reporter/...

or, the [GoConvey](http://goconvey.co/) webGUI:

    ./bin/goconvey -workDir `pwd`/src/github.com/izakmarais -excludedDirs `pwd`/src/github.com/izakmarais/reporter/tmp/