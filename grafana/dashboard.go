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
	"net/url"
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
	Title          string
	Description    string
	VariableValues string
	Rows           []Row
	Panels         []Panel
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
func NewDashboard(dashJSON []byte, variables url.Values) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}

	return dash.NewDashboard(variables)
}

func (dc dashContainer) NewDashboard(variables url.Values) Dashboard {
	var dash Dashboard
	dash.Title = sanitizeLaTexInput(dc.Dashboard.Title)
	dash.Description = sanitizeLaTexInput(dc.Dashboard.Description)
	dash.VariableValues = sanitizeLaTexInput(getVariablesValues(variables))

// Maybe some copy is not fully useful, but better be safe than sorry
// Otherwise Panel Titles were not sanitized ...
	for _, row := range dc.Dashboard.Rows {
		row.Title = sanitizeLaTexInput(row.Title)
		dash.Rows = append(dash.Rows, row)
		for _, p := range row.Panels {
			p.Title = sanitizeLaTexInput(row.Title)
			dash.Panels = append(dash.Panels, p)
		}
		dash.Rows = append(dash.Rows, lrow)
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

func expandTitleVar(input string, r *http.Request) string {
	if r == nil {
		return input
	}

	q := r.URL.Query()
//	log.Println("=======================")
	for k, v := range q {
//		log.Printf("%s -> %s\n", k, v)
		if strings.Contains(k, "var-") {
			vname := strings.Split( k, "var-")[1]
			vname1 := "$" + vname
//			log.Println("VNAME: ", vname1)
			if strings.Contains(input, vname1) {
//				log.Println("Replacing:", input, vname1, "-->", v )
				input = strings.Replace(input, vname1, strings.Join(v," "), -1 )
//				log.Println("Replacing:", input )
			}
			vname2 := "[[" + vname + "]]"
			if strings.Contains(input, vname2) {
				input = strings.Replace(input, vname2, strings.Join(v," "), -1 )
			}
		}
	}
	return input
}

func getVariablesValues(variables url.Values) string {
	values := []string{}
	for _, v := range variables {
		values = append(values, strings.Join(v, ", "))
	}
	return strings.Join(values, ", ")
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
