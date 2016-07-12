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
)

type Panel struct {
	Id   int
	Type string
}

type Dashboard struct {
	Title  string
	Panels []Panel
}

type dashContainer struct {
	Dashboard struct {
		Title string
		Rows  []struct {
			Panels []Panel
		}
	}
	Meta struct {
		Slug string
	}
}

func NewDashboard(dashJSON []byte) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)

	if err != nil {
		panic(err)
	}

	return dash.NewDashboard()
}

func (dc dashContainer) NewDashboard() Dashboard {
	var dash Dashboard
	dash.Title = dc.Dashboard.Title

	for _, row := range dc.Dashboard.Rows {
		for _, p := range row.Panels {
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
