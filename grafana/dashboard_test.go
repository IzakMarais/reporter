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
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const v4DashJSON = `
{"Dashboard":
	{"Rows":
		[{"Panels":
			[{"Type":"singlestat", "Id":1},
			 {"Type":"graph", "Id":2}]
		},
		{"Panels":
			[{"Type":"singlestat", "Id":3}]
		}]
	},
"Meta":
	{"Slug":"testDash"}
}`

const v5DashJSON = `
{"Dashboard":
	{"Panels":
		[{"Type":"singlestat", "Id":1},
		 {"Type":"graph", "Id":2},
		 {"Type":"singlestat", "Id":3}]
	},
"Meta":
	{"Slug":"testDash"}
}`

func TestV4Dashboard(t *testing.T) {
	Convey("When creating a new dashboard from Grafana v4 dashboard JSON", t, func() {
		dash := NewDashboard([]byte(v4DashJSON), url.Values{})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(dash.Panels[0].IsSingleStat(), ShouldBeTrue)
			So(dash.Panels[1].IsSingleStat(), ShouldBeFalse)
			So(dash.Panels[2].IsSingleStat(), ShouldBeTrue)
		})

		Convey("Panels should contain all panels from all rows", func() {
			So(dash.Panels, ShouldHaveLength, 3)
		})
	})
}

func TestV5Dashboard(t *testing.T) {
	Convey("When creating a new dashboard from Grafana v5 dashboard JSON", t, func() {
		dash := NewDashboard([]byte(v5DashJSON), url.Values{})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(dash.Panels[0].IsSingleStat(), ShouldBeTrue)
			So(dash.Panels[1].IsSingleStat(), ShouldBeFalse)
			So(dash.Panels[2].IsSingleStat(), ShouldBeTrue)
		})

		Convey("Panels should contain all panels from all rows", func() {
			So(dash.Panels, ShouldHaveLength, 3)
		})
	})
}
