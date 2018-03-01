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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGrafanaClientFetchesDashboard(t *testing.T) {
	Convey("When fetching a Dashboard", t, func() {
		requestURI := ""
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			fmt.Fprintln(w, `{"":""}`)
		}))
		defer ts.Close()

		grf := NewClient(ts.URL, "", url.Values{})
		grf.GetDashboard("testDash")

		Convey("It should use the dashboards endpoint", func() {
			So(requestURI, ShouldEqual, "/api/dashboards/db/testDash")
		})
	})
}

func TestGrafanaClientFetchesPanelPNG(t *testing.T) {
	Convey("When fetching a panel PNG", t, func() {
		requestURI := ""
		requestHeaders := http.Header{}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			requestHeaders = r.Header
		}))
		defer ts.Close()

		apiToken := "1234"
		variables := url.Values{}
		variables.Add("var-host", "servername")
		variables.Add("var-port", "adapter")
		grf := NewClient(ts.URL, apiToken, variables)

		grf.GetPanelPng(Panel{44, "singlestat", "title"}, "testDash", TimeRange{"now-1h", "now"})

		Convey("It should use the render endpoint with the dashboard name", func() {
			So(requestURI, ShouldStartWith, "/render/dashboard-solo/db/testDash")
		})

		Convey("It should request the panel ID", func() {
			So(requestURI, ShouldContainSubstring, "panelId=44")
		})

		Convey("It should request the time", func() {
			So(requestURI, ShouldContainSubstring, "from=now-1h")
			So(requestURI, ShouldContainSubstring, "to=now")
		})

		Convey("Singlestat panels should request a smaller size", func() {
			So(requestURI, ShouldContainSubstring, "width=300")
			So(requestURI, ShouldContainSubstring, "height=150")
		})

		Convey("apiToken should be in request header", func() {
			So(requestHeaders.Get("Authorization"), ShouldContainSubstring, apiToken)
		})

		Convey("variables should be in the request parameters", func() {
			So(requestURI, ShouldContainSubstring, "var-host=servername")
			So(requestURI, ShouldContainSubstring, "var-port=adapter")
		})

		Convey("Other panels request a larger size", func() {
			grf.GetPanelPng(Panel{44, "graph", "title"}, "testDash", TimeRange{"now", "now-1h"})
			So(requestURI, ShouldContainSubstring, "width=1000")
			So(requestURI, ShouldContainSubstring, "height=500")
		})

	})
}

func init() {
	getPanelRetrySleepTime = time.Duration(1) * time.Millisecond //we want our tests to run fast
}

func TestGrafanaClientFetchPanelPNGErrorHandling(t *testing.T) {
	Convey("When trying to fetching a panel from the server sometimes returns an error", t, func() {
		try := 0

		//create a server that will return error on the first call
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if try < 1 {
				w.WriteHeader(http.StatusInternalServerError)
				try++
			}
		}))
		defer ts.Close()

		grf := NewClient(ts.URL, "", url.Values{})

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title"}, "testDash", TimeRange{"now-1h", "now"})

		Convey("It should retry a couple of times if it receives errors", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When trying to fetching a panel from the server consistently returns an error", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		grf := NewClient(ts.URL, "", url.Values{})

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title"}, "testDash", TimeRange{"now-1h", "now"})

		Convey("The Grafana API should return an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}
