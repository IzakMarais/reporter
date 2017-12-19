
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

    go get github.com/izakmarais/reporter/...

Build and install:

    go install -v github.com/izakmarais/reporter/cmd/grafana-reporter

Running without any flags assumes Grafana is reachable at _localhost:3000_:

    grafana-reporter

Query available flags:

    grafana-reporter --help

### Generate dashboard

#### Endpoint

The reporter serves a pdf report on the specified port at:

    /api/report/{dashBoardName}

where _dashBoardName_ is the same name as used in the Grafana dashboard's URL. 
E.g. _backend-dashboard_ from _http://grafana-host:3000/dashboard/db/backend-dashboard_.

#### Query parameters

**Time span** : In addition, the endpoint supports the same time query parameters as Grafana. 
This means that you can create a Grafana Link and enable the _Time range_ forwarding check-box.
The link will render a dashboard with your current dashboard time range.

**template**: Optionally specify a custom TeX template file.
 _template=templateName_ implies a template file at `templates/templateName.tex`.
 The `templates` directory can be set with a commandline parameter.
 Sample extended template (enhanced.tex) is provided that utilizes all the reporter features (might require additional LaTeX modules to work)   

 **apitoken**: Optionally specify a Grafana authentication api token. Use this if you have auth enabled on Grafana.

 **variable**: Allows to pass single variable to Grafana, in standard format compatible with dashboard links (i.e: &var-VariableName=VariableValue)

### Test

The unit tests can be run using the go tool:

    go get github.com/smartystreets/goconvey
    go test -v github.com/izakmarais/reporter/...

or, the [GoConvey](http://goconvey.co/) webGUI:

    ./bin/goconvey -workDir `pwd`/src/github.com/izakmarais -excludedDirs `pwd`/src/github.com/izakmarais/reporter/tmp/
