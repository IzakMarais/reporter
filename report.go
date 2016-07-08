/*
   Copyright 2013 Vastech SA (PTY) LTD

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
func (this *Report) Generate() *os.File {
	dash := this.gClient.GetDashboard(this.dashName)
	this.renderPNGsParallel(dash)
	this.generateTeXFile(dash)
	return this.runLaTeX()
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

func (this *Report) renderPNGsParallel(dash grafana.Dashboard) {
	var wg sync.WaitGroup
	wg.Add(len(dash.Panels))

	for _, p := range dash.Panels {
		go func(p grafana.Panel) {
			defer wg.Done()
			this.renderPNG(p)
		}(p)
	}

	wg.Wait()
}

func (this *Report) renderPNG(p grafana.Panel) {
	body := this.gClient.GetPanelPng(p, this.dashName, this.time)
	defer body.Close()

	err := os.MkdirAll(this.imgDirPath(), 0777)
	stopIf(err)
	imgFileName := fmt.Sprintf("image%d.png", p.Id)
	file, err := os.Create(filepath.Join(this.imgDirPath(), imgFileName))
	stopIf(err)
	defer file.Close()

	_, err = io.Copy(file, body)
	stopIf(err)
}

func (this *Report) generateTeXFile(dash grafana.Dashboard) {
	type templData struct {
		grafana.Dashboard
		grafana.TimeRange
	}

	err := os.MkdirAll(this.tmpDir, 0777)
	stopIf(err)
	file, err := os.Create(this.texPath())
	stopIf(err)
	defer file.Close()

	tmpl := template.Must(
		template.New("report").Delims("[[", "]]").Parse(texTemplate))
	data := templData{dash, this.time}
	err = tmpl.Execute(file, data)
	stopIf(err)
}

func (this *Report) runLaTeX() *os.File {
	cmd := exec.Command("pdflatex", "-halt-on-error", reportTexFile)
	cmd.Dir = this.tmpDir
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		stopIf(errors.New("Latex failed with output: " + string(outBytes)))
	}

	file, err := os.Open(this.pdfPath())
	stopIf(err)
	return file
}

func stopIf(err error) {
	if err != nil {
		panic(err)
	}
}
