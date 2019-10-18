/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/IzakMarais/reporter/grafana"
	"github.com/IzakMarais/reporter/report"
	"github.com/gorilla/mux"
)

var proto = flag.String("proto", "http://", "Grafana Protocol. Change to 'https://' if Grafana is using https. Reporter will still serve http.")
var ip = flag.String("ip", "localhost:3000", "Grafana IP and port.")
var port = flag.String("port", ":8686", "Port to serve on.")
var templateDir = flag.String("templates", "templates/", "Directory for custom TeX templates.")
var sslCheck = flag.Bool("ssl-check", true, "Check the SSL issuer and validity. Set this to false if your Grafana serves https using an unverified, self-signed certificate.")
var gridLayout = flag.Bool("grid-layout", false, "Enable grid layout (-grid-layout=1). Panel width and height will be calculated based off Grafana gridPos width and height.")

//cmd line mode params
var cmdMode = flag.Bool("cmd_enable", false, "Enable command line mode. Generate report from command line without starting webserver (-cmd_enable=1).")
var dashboard = flag.String("cmd_dashboard", "", "Dashboard identifier. Required (and only used) in command line mode.")
var apiKey = flag.String("cmd_apiKey", "", "Grafana api key. Required (and only used) in command line mode.")
var apiVersion = flag.String("cmd_apiVersion", "v5", "Api version: [v4, v5]. Required (and only used) in command line mode, example: -apiVersion v5.")
var outputFile = flag.String("cmd_o", "out.pdf", "Output file. Required (and only used) in command line mode.")
var timeSpan = flag.String("cmd_ts", "from=now-3h&to=now", "Time span. Required (and only used) in command line mode.")
var template = flag.String("cmd_template", "", "Specify a custom TeX template file. Only used in command line mode, but is optional even there.")

func main() {
	flag.Parse()
	log.SetOutput(os.Stdout)

	//'generated*'' variables injected from build.gradle: task 'injectGoVersion()'
	log.Printf("grafana reporter, version: %s.%s-%s hash: %s", generatedMajor, generatedMinor, generatedRelease, generatedGitHash)
	log.Printf("serving at '%s' and using grafana at '%s'", *port, *proto+*ip)
	if !*sslCheck {
		log.Printf("SSL check disabled")
	} else {
		log.Printf("SSL check enforced")
	}
	if !*gridLayout {
		log.Printf("Using sequential report layout. Consider enabling 'grid-layout' so that your report more closely follow the dashboard layout.")
	} else {
		log.Printf("Using grid layout.")
	}

	router := mux.NewRouter()
	RegisterHandlers(
		router,
		ServeReportHandler{grafana.NewV4Client, report.New},
		ServeReportHandler{grafana.NewV5Client, report.New},
	)

	if *cmdMode {
		log.Printf("Called with command line mode enabled, will save report to file and exit.")
		log.Printf("Called with command line mode 'dashboard' '%s'", *dashboard)
		log.Printf("Called with command line mode 'apiKey' '%s'", *apiKey)
		log.Printf("Called with command line mode 'apiVersion' '%s'", *apiVersion)
		log.Printf("Called with command line mode 'outputFile' '%s'", *outputFile)
		log.Printf("Called with command line mode 'timeSpan' '%s'", *timeSpan)
		if template != nil && *template != "" {
			log.Printf("Called with command line mode 'template' '%s'", *template)
		}

		if err := cmdHandler(router); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatal(http.ListenAndServe(*port, router))
	}
}
