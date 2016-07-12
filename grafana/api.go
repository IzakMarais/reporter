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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"errors"
)

type Client interface {
	GetDashboard(dashName string) Dashboard
	GetPanelPng(p Panel, dashName string, t TimeRange) io.ReadCloser
}

type client struct {
	url string
	apiToken string
}

func NewClient(url string, apiToken string) Client {
	return client{url, apiToken}
}

func (g client) GetDashboard(dashName string) Dashboard {
	dashURL := g.url + "/api/dashboards/db/" + dashName
	log.Println("Connecting to dashboard at", dashURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", dashURL, nil)
	if err != nil {
		panic(err)
	}

	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer " + g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		log.Println(string(body))
		panic("Error obtaining dashboard")
	}

	return NewDashboard(body)
}

func (g client) GetPanelPng(p Panel, dashName string, t TimeRange) io.ReadCloser {

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

	panelUrl := fmt.Sprintf("%s/render/dashboard-solo/db/%s?%s",
		g.url, dashName, v.Encode())

	log.Println("Downloading image ", p.Id, panelUrl)

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("Failed to login. Need API Token")
	}}
	req, err := http.NewRequest("GET", panelUrl, nil)
	if err != nil {
		panic(err)
	}
	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer " + g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))
		panic("Error obtaining render")
	}

	return resp.Body
}
