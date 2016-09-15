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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/izakmarais/reporter/grafana"
	"github.com/pborman/uuid"
	"io/ioutil"
)

type Report struct {
	gClient  grafana.Client
	time     grafana.TimeRange
	dashName string
	tmpDir   string
}

const (
	imgDir        = "images"
	reportTexFile = "report.tex"
	reportPdf     = "report.pdf"
)

func New(g grafana.Client, dashName string, time grafana.TimeRange) Report {
	tmpDir := filepath.Join("tmp", uuid.New())
	return Report{g, time, dashName, tmpDir}
}

// Generate returns the report.pdf file.  After reading this file it should be Closed()
// After closing the file, call report.Clean() to delete the file as well the temporary build files
func (this *Report) Generate(templateFile string) (latex_file *os.File, err error) {
	dash,err := this.gClient.GetDashboard(this.dashName)
	if err != nil {
		return
	}
	err = this.renderPNGsParallel(dash)
	if err != nil {
		return
	}
	err = this.generateTeXFile(dash, templateFile)
	if err != nil {
		return
	}
	latex_file,err = this.runLaTeX()
	return
}

func (this *Report) Clean() {
	err := os.RemoveAll(this.tmpDir)
	if err != nil {
		log.Println("Error cleaning up tmp dir:", err)
	}
}

func (this *Report) imgDirPath() string {
	return filepath.Join(this.tmpDir, imgDir)
}

func (this *Report) pdfPath() string {
	return filepath.Join(this.tmpDir, reportPdf)
}

func (this *Report) texPath() string {
	return filepath.Join(this.tmpDir, reportTexFile)
}

func (this *Report) renderPNGsParallel(dash grafana.Dashboard) (err error){
	var wg sync.WaitGroup
	wg.Add(len(dash.Panels))

	for _, p := range dash.Panels {
		go func(p grafana.Panel) {
			defer wg.Done()
			err = this.renderPNG(p)
			if err != nil {
				return
			}
		}(p)
	}

	wg.Wait()
	return
}

func (this *Report) renderPNG(p grafana.Panel) (err error) {
	body,err := this.gClient.GetPanelPng(p, this.dashName, this.time)
	if err != nil {
		return
	}
	defer body.Close()

	err = os.MkdirAll(this.imgDirPath(), 0777)
	if err != nil {
		return
	}
	imgFileName := fmt.Sprintf("image%d.png", p.Id)
	file, err := os.Create(filepath.Join(this.imgDirPath(), imgFileName))
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	return
}

func (this *Report) generateTeXFile(dash grafana.Dashboard, templateFile string) (err error) {
	type templData struct {
		grafana.Dashboard
		grafana.TimeRange
	}

	err = os.MkdirAll(this.tmpDir, 0777)
	if err != nil {
		return
	}
	file, err := os.Create(this.texPath())
	if err != nil {
		return
	}
	defer file.Close()


	texTemplate,err := ioutil.ReadFile(templateFile)
	if err != nil {
		err = errors.New("Error reading template: " + err.Error())
		return
	}

	tmpl := template.Must(
		template.New("report").Delims("[[", "]]").Parse(string(texTemplate)))
	data := templData{dash, this.time}
	err = tmpl.Execute(file, data)
	return
}

func (this *Report) runLaTeX() (latex_file *os.File, err error) {
	cmd := exec.Command("pdflatex", "-halt-on-error", reportTexFile)
	cmd.Dir = this.tmpDir
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New("Latex failed with output: " + string(outBytes))
		return
	}

	latex_file, err = os.Open(this.pdfPath())
	return
}

