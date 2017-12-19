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
	"encoding/json"
	"strings"
)

// Panel represents a Grafana dashboard panel
type Panel struct {
	Id    int
	Type  string
	Title string
}

// Row represents a container for Panels
type Row struct {
	Id        int
	Showtitle bool
	Title     string
	Panels    []Panel
}

// Dashboard represents a Grafana dashboard
type Dashboard struct {
	Title       string
	Description string
	Variable    string
	Rows        []Row
	Panels      []Panel
}

type dashContainer struct {
	Dashboard struct {
		Title       string
		Description string
		Rows        []Row
	}
	Meta struct {
		Slug string
	}
}

// NewDashboard creates Dashboard from Grafana's internal JSON dashboard definition
func NewDashboard(dashJSON []byte, variable string) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	return dash.NewDashboard(variable)
}

func (dc dashContainer) NewDashboard(variable string) Dashboard {
	var dash Dashboard
	dash.Title = sanitizeLaTexInput(dc.Dashboard.Title)
	dash.Description = sanitizeLaTexInput(dc.Dashboard.Description)
	dash.Variable = sanitizeLaTexInput(variable)

	for _, row := range dc.Dashboard.Rows {
		row.Title = sanitizeLaTexInput(row.Title)
		dash.Rows = append(dash.Rows, row)
		for _, p := range row.Panels {
			p.Title = sanitizeLaTexInput(row.Title)
			dash.Panels = append(dash.Panels, p)
		}
	}

	return dash
}

func (p Panel) IsSingleStat() bool {
	if p.Type == "singlestat" {
		return true
	}
	return false
}

func (r Row) IsVisible() bool {
	return r.Showtitle
}

func (d Dashboard) GetVariable() string {
	if strings.Contains(d.Variable, "=") {
		return strings.Split(d.Variable, "=")[1]
	}
	return "-"
}

func sanitizeLaTexInput(input string) string {
	input = strings.Replace(input, "\\", "\\textbackslash ", -1)
	input = strings.Replace(input, "&", "\\&", -1)
	input = strings.Replace(input, "%", "\\%", -1)
	input = strings.Replace(input, "$", "\\$", -1)
	input = strings.Replace(input, "#", "\\#", -1)
	input = strings.Replace(input, "_", "\\_", -1)
	input = strings.Replace(input, "{", "\\{", -1)
	input = strings.Replace(input, "}", "\\}", -1)
	input = strings.Replace(input, "~", "\\textasciitilde ", -1)
	input = strings.Replace(input, "^", "\\textasciicircum ", -1)
	return input
}
