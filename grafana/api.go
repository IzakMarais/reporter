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
	url       string
	apiToken  string
	variables url.Values
}

// NewClient creates a new Grafana Client. If apiToken is the empty string,
// authorization headers will be omitted from requests.
// variables are Grafana template variable url values of the form var-{name}={value}, e.g. var-host=dev
func NewClient(url string, apiToken string, variables url.Values) Client {
	return client{url, apiToken, variables}
}

func (g client) GetDashboard(dashName string) (dashboard Dashboard, err error) {
	dashURL := g.url + "/api/dashboards/db/" + dashName
	if len(g.variables) > 0 {
		dashURL = dashURL + "?" + g.variables.Encode()
	}
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

	dashboard = NewDashboard(body, g.variables)
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
	values := url.Values{}
	values.Add("theme", "light")
	values.Add("panelId", strconv.Itoa(p.Id))
	values.Add("from", t.From)
	values.Add("to", t.To)
	if p.IsSingleStat() {
		values.Add("width", "300")
		values.Add("height", "150")
	} else {
		values.Add("width", "1000")
		values.Add("height", "500")
	}

	for k, v := range g.variables {
		for _, singleValue := range v {
			values.Add(k, singleValue)
		}
	}

	url := fmt.Sprintf("%s/render/dashboard-solo/db/%s?%s", g.url, dashName, values.Encode())
	log.Println("Downloading image ", p.Id, url)
	return url
}
