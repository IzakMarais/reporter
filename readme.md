
# Grafana reporter <img style="float: right;" src="https://travis-ci.org/IzakMarais/reporter.svg?branch=master">

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

    go get github.com/IzakMarais/reporter/...

Build and install `grafana-reporter` binary to `$GOPATH/bin`:

    go install -v github.com/IzakMarais/reporter/cmd/grafana-reporter

Running without any flags assumes Grafana is reachable at `localhost:3000`:

    grafana-reporter

Query available flags:

    grafana-reporter --help

#### Docker examples (optional)

If you also have `Make`, `Docker` and `Docker-compose` installed, you can do the following.
First change directory :

    cd github.com/IzakMarais/reporter

Generate a docker image including `pdflatex` (warning: the TeXLive install can take a long time):

    make docker-build

To run a simple local orchestration of Grafana and Grafana-Reporter:

     go get github.com/IzakMarais/reporter/ ...
     cd $GOPATH/src/github.com/IzakMarais/reporter
     make compose-up

Then open a browser to `http://localhost:3000` and create a new test dashboard. Add the example graph and save the dashboard as `test`.
Next, open another browser window/tab and go to: `http://localhost:8080/api/report/test` which will output the grafana-reporter PDF.

### Generate a dashboard report

#### Endpoint

The reporter serves a pdf report on the specified port at:

    /api/report/{dashBoardName}

where `{dashBoardName}` is the same name as used in the Grafana dashboard's URL.
E.g. `backend-dashboard` from `http://grafana-host:3000/dashboard/db/backend-dashboard`.

#### Query parameters

**Time span** : In addition, the endpoint supports the same time query parameters as Grafana.
This means that you can create a Grafana Link and enable the _Time range_ forwarding check-box.
The link will render a dashboard with your current dashboard time range.

**template**: Optionally specify a custom TeX template file.
 `template=templateName` implies a template file at `templates/templateName.tex`.
 The `templates` directory can be set with a commandline parameter.

 **apitoken**: Optionally specify a Grafana authentication api token. Use this if you have auth enabled on Grafana. Example:

    /api/report/{dashBoardName}&apitoken={tokenstring}

where `{dashBoardName}` and `{tokenstring}` should be substituted with your desired values.

## Test

The unit tests can be run using the go tool:

    go test -v github.com/IzakMarais/reporter/...

or, the [GoConvey](http://goconvey.co/) webGUI:

    ./bin/goconvey -workDir `pwd`/src/github.com/IzakMarais -excludedDirs `pwd`/src/github.com/IzakMarais/reporter/tmp/
