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
	"net/http"
//	"log"
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
        Rows	    []Row
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
	var lrow Row
	var lpan Panel

	dash.Title = sanitizeLaTexInput(dc.Dashboard.Title)
	dash.Description = sanitizeLaTexInput(dc.Dashboard.Description)
	dash.Variable = sanitizeLaTexInput(variable)

// Maybe some copy is not fully useful, but better be safe than sorry
// Otherwise Panel Titles were not sanitized ...
	for _, row := range dc.Dashboard.Rows {
		lrow = row
		lrow.Panels = nil
		lrow.Title = expandTitleVar(lrow.Title, GlobalReq)
		lrow.Title = sanitizeLaTexInput(lrow.Title)
		for _, p := range row.Panels {
			lpan = p
			lpan.Title = expandTitleVar(lpan.Title, GlobalReq)
			lpan.Title = sanitizeLaTexInput(lpan.Title)
			lrow.Panels = append(lrow.Panels, lpan)
			dash.Panels = append(dash.Panels, lpan)
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

func (d Dashboard) GetVariable() string {
	if strings.Contains(d.Variable, "=") {
		return strings.Split(d.Variable, "=")[1]
	}
	return "-"
}

func expandTitleVar(input string, r *http.Request) string {
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

func sanitizeLaTexInput(input string) string {
//log.Println("sanitizeLaTexInput  IN ",input)
	input = strings.Replace(input, "\\", "\\textbackslash ", -1)
	input = strings.Replace(input, "&", "\\&", -1)
	input = strings.Replace(input, "%", "\\%", -1)
	input = strings.Replace(input, "$", "\\textdollar ", -1)
	input = strings.Replace(input, "#", "\\#", -1)
	input = strings.Replace(input, "_", "\\_", -1)
	input = strings.Replace(input, "{", "\\{", -1)
	input = strings.Replace(input, "}", "\\}", -1)
	input = strings.Replace(input, "~", "\\textasciitilde ", -1)
	input = strings.Replace(input, "^", "\\textasciicircum ", -1)
//log.Println("sanitizeLaTexInput OUT ",input)
	return input
}

