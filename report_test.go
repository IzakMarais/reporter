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

package report

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/izakmarais/reporter/grafana"
	. "github.com/smartystreets/goconvey/convey"
)

const dashJSON = `
{"Dashboard":
	{
		"Title":"My first dashboard",
		"Rows":
		[{"Panels":
			[{"Type":"singlestat", "Id":1},
			 {"Type":"graph", "Id":22}]
		},
		{"Panels":
			[{"Type":"singlestat", "Id":33}]
		}]
	},
"Meta":
	{"Slug":"testDash"}
}`

type mockGrafanaClient struct {
}

func (m mockGrafanaClient) GetDashboard(dashName string) (grafana.Dashboard, error) {
	return grafana.NewDashboard([]byte(dashJSON)), nil
}

func (m mockGrafanaClient) GetPanelPng(p grafana.Panel, dashName string, t grafana.TimeRange) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer([]byte("Not actually a png"))), nil
}

func TestReport(t *testing.T) {
	Convey("When generating a report", t, func() {
		var gClient mockGrafanaClient
		rep := New(gClient, "testDash", grafana.TimeRange{"1453206447000", "1453213647000"}, "")
		defer rep.Clean()

		Convey("When rendering images", func() {
			dashboard, _ := gClient.GetDashboard("")
			rep.renderPNGsParallel(dashboard)

			Convey("It should create a temporary folder", func() {
				_, err := os.Stat(rep.tmpDir)
				So(err, ShouldBeNil)
			})

			Convey("It should copy the file to the image folder", func() {
				_, err := os.Stat(rep.imgDirPath() + "/image1.png")
				So(err, ShouldBeNil)
			})

			Convey("It should create one file per panel", func() {
				f, err := os.Open(rep.imgDirPath())
				defer f.Close()
				files, err := f.Readdir(0)
				So(files, ShouldHaveLength, 3)
				So(err, ShouldBeNil)
			})
		})

		Convey("When genereting the Tex file", func() {
			dashboard, _ := gClient.GetDashboard("")
			rep.generateTeXFile(dashboard)
			f, err := os.Open(rep.texPath())
			defer f.Close()

			Convey("It should create a file in the temporary folder", func() {
				So(err, ShouldBeNil)
			})

			Convey("The file should contain reference to the template data", func() {
				var buf bytes.Buffer
				io.Copy(&buf, f)
				s := string(buf.Bytes())

				So(err, ShouldBeNil)
				Convey("Including the Title", func() {
					So(s, ShouldContainSubstring, "My first dashboard")

				})
				Convey("and the images", func() {
					So(s, ShouldContainSubstring, "image1")
					So(s, ShouldContainSubstring, "image22")
					So(s, ShouldContainSubstring, "image33")

				})
				Convey("and the time range", func() {
					So(s, ShouldContainSubstring, "Tue Jan 19 12:27:27 UTC 2016")
					So(s, ShouldContainSubstring, "Tue Jan 19 14:27:27 UTC 2016")
				})
			})
		})

		Convey("Clean() should remove the temporary folder", func() {
			rep.Clean()

			_, err := os.Stat(rep.tmpDir)
			So(os.IsNotExist(err), ShouldBeTrue)
		})
	})

}
