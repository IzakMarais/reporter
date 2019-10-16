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
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/IzakMarais/reporter/grafana"
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
			[
				{"Type":"singlestat", "Id":33},
				{"Type":"graph", "Id":44},
				{"Type":"graph", "Id":55},
				{"Type":"graph", "Id":66},
				{"Type":"graph", "Id":77},
				{"Type":"graph", "Id":88},
				{"Type":"graph", "Id":99}
			]
		}]
	},
"Meta":
	{"Slug":"testDash"}
}`

type mockGrafanaClient struct {
	getPanelCallCount int
	variables         url.Values
}

func (m *mockGrafanaClient) GetDashboard(dashName string) (grafana.Dashboard, error) {
	return grafana.NewDashboard([]byte(dashJSON), m.variables), nil
}

func (m *mockGrafanaClient) GetPanelPng(p grafana.Panel, dashName string, t grafana.TimeRange) (io.ReadCloser, error) {
	m.getPanelCallCount++
	return ioutil.NopCloser(bytes.NewBuffer([]byte("Not actually a png"))), nil
}

func TestReport(t *testing.T) {
	Convey("When generating a report", t, func() {
		variables := url.Values{}
		variables.Add("var-test", "testvarvalue")
		gClient := &mockGrafanaClient{0, variables}
		rep := new(gClient, "testDash", grafana.TimeRange{"1453206447000", "1453213647000"}, "", false)
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

			Convey("It shoud call getPanelPng once per panel", func() {
				So(gClient.getPanelCallCount, ShouldEqual, 9)
			})

			Convey("It should create one file per panel", func() {
				f, err := os.Open(rep.imgDirPath())
				defer f.Close()
				files, err := f.Readdir(0)
				So(files, ShouldHaveLength, 9)
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
				Convey("Including the varialbe values", func() {
					So(s, ShouldContainSubstring, "testvarvalue")

				})
				Convey("and the images", func() {
					So(s, ShouldContainSubstring, "image1")
					So(s, ShouldContainSubstring, "image22")
					So(s, ShouldContainSubstring, "image33")
					So(s, ShouldContainSubstring, "image44")
					So(s, ShouldContainSubstring, "image55")
					So(s, ShouldContainSubstring, "image66")
					So(s, ShouldContainSubstring, "image77")
					So(s, ShouldContainSubstring, "image88")
					So(s, ShouldContainSubstring, "image99")
				})
				Convey("and the time range", func() {
					//server time zone by shift hours timestamp
					//so just test for day and year
					So(s, ShouldContainSubstring, "Tue Jan 19")
					So(s, ShouldContainSubstring, "2016")
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

type errClient struct {
	getPanelCallCount int
	variables         url.Values
}

func (e *errClient) GetDashboard(dashName string) (grafana.Dashboard, error) {
	return grafana.NewDashboard([]byte(dashJSON), e.variables), nil
}

//Produce an error on the 2nd panel fetched
func (e *errClient) GetPanelPng(p grafana.Panel, dashName string, t grafana.TimeRange) (io.ReadCloser, error) {
	e.getPanelCallCount++
	if e.getPanelCallCount == 2 {
		return nil, errors.New("The second panel has some problem")
	}
	return ioutil.NopCloser(bytes.NewBuffer([]byte("Not actually a png"))), nil
}

func TestReportErrorHandling(t *testing.T) {
	Convey("When generating a report where one panels gives an error", t, func() {
		variables := url.Values{}
		gClient := &errClient{0, variables}
		rep := new(gClient, "testDash", grafana.TimeRange{"1453206447000", "1453213647000"}, "", false)
		defer rep.Clean()

		Convey("When rendering images", func() {
			dashboard, _ := gClient.GetDashboard("")
			err := rep.renderPNGsParallel(dashboard)

			Convey("It shoud call getPanelPng once per panel", func() {
				So(gClient.getPanelCallCount, ShouldEqual, 9)
			})

			Convey("It should create one less image file than the total number of panels", func() {
				f, err := os.Open(rep.imgDirPath())
				defer f.Close()
				files, err := f.Readdir(0)
				So(files, ShouldHaveLength, 8) //one less than the total number of im
				So(err, ShouldBeNil)
			})

			Convey("If any panels return errors, renderPNGsParralel should return the error message from one panel", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "The second panel has some problem")
			})
		})

		Convey("Clean() should remove the temporary folder", func() {
			rep.Clean()

			_, err := os.Stat(rep.tmpDir)
			So(os.IsNotExist(err), ShouldBeTrue)
		})
	})

}
