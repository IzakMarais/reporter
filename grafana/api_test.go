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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGrafanaClient(t *testing.T) {

	Convey("When fetching a Dashboard", t, func() {
		requestURI := ""
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			fmt.Fprintln(w, `{"":""}`)
		}))
		defer ts.Close()

		api_token := ""
		grf := NewClient(ts.URL, api_token)
		grf.GetDashboard("testDash")

		Convey("It should use the dashboards endpoint", func() {
			So(requestURI, ShouldEqual, "/api/dashboards/db/testDash")
		})
	})

	Convey("When fetching a panel PNG", t, func() {
		requestURI := ""

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
		}))
		defer ts.Close()

		api_token := ""
		grf := NewClient(ts.URL, api_token)

		grf.GetPanelPng(Panel{44, "singlestat"}, "testDash", TimeRange{"now-1h", "now"})

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

		Convey("Other panels request a larger size", func() {
			grf.GetPanelPng(Panel{44, "graph"}, "testDash", TimeRange{"now", "now-1h"})
			So(requestURI, ShouldContainSubstring, "width=1000")
			So(requestURI, ShouldContainSubstring, "height=500")
		})

	})

}
