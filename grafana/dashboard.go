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
	"log"
	"net/url"
	"strings"
	"fmt"
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
// This is both used to unmarshal the dashbaord JSON into
// and then enriched (sanitize fields for TeX consumption and add VarialbeValues)
type Dashboard struct {
	Title          string
	Description    string
	VariableValues string //Not present in the Grafana JSON structure. Enriched data passed used by the Tex templating
	Rows           []Row
	Panels         []Panel
	VArray         url.Values   //Not present in the Grafana JSON structure. Enriched data passed used by the Tex templating
}

type dashContainer struct {
	Dashboard Dashboard
	Meta      struct {
		Slug string
	}
}

// NewDashboard creates Dashboard from Grafana's internal JSON dashboard definition
func NewDashboard(dashJSON []byte, variables url.Values) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	d := dash.NewDashboard(variables)
	log.Printf("Populated dashboard datastructure: %+v\n", d)
	return d
}

func (dc dashContainer) NewDashboard(variables url.Values) Dashboard {
	var dash Dashboard
	dash.Title = sanitizeLaTexInput(dc.Dashboard.Title)
	dash.Description = sanitizeLaTexInput(dc.Dashboard.Description)
	dash.VariableValues = sanitizeLaTexInput(getVariablesValues(variables))
	dash.VArray = getVariablesArray(variables)

	if len(dc.Dashboard.Rows) == 0 {
		return populatePanelsFromV5JSON(dash, dc)
	}
	return populatePanelsFromV4JSON(dash, dc)
}

func populatePanelsFromV4JSON(dash Dashboard, dc dashContainer) Dashboard {
/*- OLD Code -*/
	for _, row := range dc.Dashboard.Rows {
		row.Title = sanitizeLaTexInput(row.Title)
		for i, p := range row.Panels {
			p.Title = sanitizeLaTexInput(p.Title)
			row.Panels[i] = p
			dash.Panels = append(dash.Panels, p)
		}
		dash.Rows = append(dash.Rows, row)
	}
/*- -*/

/*- 
	for _, row := range dc.Dashboard.Rows {
		row.Title = sanitizeLaTexInput(row.Title)
		plist := row.Panels
		row.Panels = nil
		for _, p := range plist {
			p.Title = sanitizeLaTexInput(p.Title)
			dash.Panels = append(dash.Panels, p)
			row.Panels = append(row.Panels, p)
		}
		dash.Rows = append(dash.Rows, row)
	}
 -*/

	return dash
}

func populatePanelsFromV5JSON(dash Dashboard, dc dashContainer) Dashboard {
	for _, p := range dc.Dashboard.Panels {
		if p.Type == "row" {
			continue
		}
		p.Title = sanitizeLaTexInput(p.Title)
		dash.Panels = append(dash.Panels, p)
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

func (d Dashboard) GetCleanVar(s []string) string {
	return strings.Trim(fmt.Sprintf(strings.Join(s, ", ")), "[]")
}

func getVariablesValues(variables url.Values) string {
	values := []string{}
	for _, v := range variables {
		values = append(values, strings.Join(v, ", "))
	}
	return strings.Join(values, ", ")
}

func getVariablesArray(variables url.Values) url.Values {
	values := url.Values{}
	for k, v := range variables {
		values.Add (sanitizeLaTexInput(k), sanitizeLaTexInput(strings.Join(v, ", ")))
	}
	return values
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
