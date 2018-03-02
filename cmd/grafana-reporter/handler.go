/*
   Copyright 2018 Vastech SA (PTY) LTD

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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/IzakMarais/reporter/grafana"
	"github.com/IzakMarais/reporter/report"
	"github.com/gorilla/mux"
)

// ServeReportHandler allows mocking of the grafana client and report generation
type ServeReportHandler struct {
	newGrafanaClient func(url string, apiToken string, variables url.Values) grafana.Client
	newReport        func(g grafana.Client, dashName string, time grafana.TimeRange, texTemplate string) report.Report
}

//RegisterHandlers registers all http.Handler's with their associated routes to the router
func RegisterHandlers(router *mux.Router, reportServer ServeReportHandler) {
	router.Handle("/api/report/{dashName}", reportServer)
}

func (h ServeReportHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Print("Reporter called")
	g := h.newGrafanaClient(*proto+*ip, apiToken(req), dashVariables(req))
	rep := h.newReport(g, dashName(req), time(req), texTemplate(req))

	file, err := rep.Generate()
	if err != nil {
		log.Println("Error generating report:", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rep.Clean()
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		log.Println("Error copying data to response:", err)
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println("Report generated correctly")
}

func dashName(r *http.Request) string {
	vars := mux.Vars(r)
	d := vars["dashName"]
	log.Println("Called with dashboard:", d)
	return d
}

func time(r *http.Request) grafana.TimeRange {
	params := r.URL.Query()
	t := grafana.NewTimeRange(params.Get("from"), params.Get("to"))
	log.Println("Called with time range:", t)
	return t
}

func apiToken(r *http.Request) string {
	apiToken := r.URL.Query().Get("apitoken")
	log.Println("Called with api Token:", apiToken)
	return apiToken
}

func dashVariables(r *http.Request) url.Values {
	output := url.Values{}
	for k, v := range r.URL.Query() {
		if strings.HasPrefix(k, "var-") {
			log.Println("Called with variable:", k, v)
			for _, singleV := range v {
				output.Add(k, singleV)
			}
		}
	}
	if len(output) == 0 {
		log.Println("Called without variable")
	}
	return output
}

func texTemplate(r *http.Request) string {
	fName := r.URL.Query().Get("template")
	if fName == "" {
		return ""
	}
	file := filepath.Join(*templateDir, fName+".tex")
	log.Println("Called with template:", file)

	customTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("Error reading template file: %q", err)
		return ""
	}

	return string(customTemplate)
}
