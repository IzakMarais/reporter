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

package grafana

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"net/http/httputil"
	"strings"
)

// Global var to store the HTTP Request from Main
var GlobalReq *http.Request

// Client is a Grafana API client
type Client interface {
	GetDashboard(dashName string) (Dashboard, error)
	GetPanelPng(p Panel, dashName string, t TimeRange) (io.ReadCloser, error)
}

type client struct {
	url      string
	apiToken string
	variable string
}

// NewClient creates a new Grafana Client. If apiToken is the empty string,
// authorization headers will be omitted from requests.
func NewClient(url string, apiToken string, variable string) Client {
	return client{url, apiToken, variable}
}

func (g client) GetDashboard(dashName string) (dashboard Dashboard, err error) {
	dashURL := g.url + "/api/dashboards/db/" + dashName
	log.Println("Connecting to dashboard at", dashURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", dashURL, nil)
	if err != nil {
		return
	}

	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		log.Println("Error obtaining dashboard: ", resp.Status)
		err = errors.New("Error obtaining dashboard: " + string(body))
		return
	}

	dashboard = NewDashboard(body, g.variable)
	return
}

func (g client) GetPanelPng(p Panel, dashName string, t TimeRange) (body io.ReadCloser, err error) {
	panelURL := g.getPanelURL(p, dashName, t)

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("Error getting panel png. Redirected to login")
	}}
	req, err := http.NewRequest("GET", panelURL, nil)
	if err != nil {
		return
	}
	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))
		err = errors.New("Error obtaining render: " + resp.Status)
	}

	body = resp.Body
	return
}

func (g client) getPanelURL(p Panel, dashName string, t TimeRange) string {
	var httpvars string
	v := url.Values{}
	v.Add("theme", "light")
	v.Add("panelId", strconv.Itoa(p.Id))
	v.Add("from", t.From)
	v.Add("to", t.To)
	if p.IsSingleStat() {
		v.Add("width", "300")
		v.Add("height", "150")
	} else {
		v.Add("width", "1000")
		v.Add("height", "500")
	}

	httpvars = ""
	// Save a copy of this request for debugging.
	if GlobalReq != nil {
		// Convert to bytearray : GET /api/report/DashName? ... HTTP/1.1
//		log.Println("Managing HTTP Variables ")
		requestDump, err := httputil.DumpRequest(GlobalReq, true)
		if err != nil {
	  		log.Println("Global Request - Error: ", err)
		}
		//log.Println(string(requestDump))
	
		// Grafana variables all have "var-" at the beginning 
		rini := strings.Index(string(requestDump), "&var-")
		// Request ends with " HTTP/1.1"
		rfin := strings.Index(string(requestDump), " HTTP/1.1")

		// Add the useful part of request to the URI down to the Panel, let the Panel grab what it needs
		httpvars = string(requestDump[rini:rfin])
//        	url := fmt.Sprintf("%s/render/dashboard-solo/db/%s?%s%s", g.url, dashName, v.Encode(), string(requestDump[rini:rfin]))
	}
	url := fmt.Sprintf("%s/render/dashboard-solo/db/%s?%s%s", g.url, dashName, v.Encode(), httpvars)

	log.Println("Downloading image ", p.Id, url[:90], "...")
	return url
}
