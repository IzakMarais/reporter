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
	gClient     grafana.Client
	time        grafana.TimeRange
	texTemplate string
	dashName    string
	tmpDir      string
}

const (
	imgDir        = "images"
	reportTexFile = "report.tex"
	reportPdf     = "report.pdf"
)

// New creates a new Report. If texTemplate is empty, a default tex template is used.
func New(g grafana.Client, dashName string, time grafana.TimeRange, texTemplate string) Report {
	if texTemplate == "" {
		texTemplate = defaultTemplate
	}
	tmpDir := filepath.Join("tmp", uuid.New())
	return Report{g, time, texTemplate, dashName, tmpDir}
}

// Generate returns the report.pdf file.  After reading this file it should be Closed()
// After closing the file, call report.Clean() to delete the file as well the temporary build files
func (rep *Report) Generate() (pdf *os.File, err error) {
	dash, err := rep.gClient.GetDashboard(rep.dashName)
	if err != nil {
		return
	}
	err = rep.renderPNGsParallel(dash)
	if err != nil {
		return
	}
	err = rep.generateTeXFile(dash)
	if err != nil {
		return
	}
	pdf, err = rep.runLaTeX()
	return
}

// Clean deletes the temporary directory used during report generation
func (rep *Report) Clean() {
	err := os.RemoveAll(rep.tmpDir)
	if err != nil {
		log.Println("Error cleaning up tmp dir:", err)
	}
}

func (rep *Report) imgDirPath() string {
	return filepath.Join(rep.tmpDir, imgDir)
}

func (rep *Report) pdfPath() string {
	return filepath.Join(rep.tmpDir, reportPdf)
}

func (rep *Report) texPath() string {
	return filepath.Join(rep.tmpDir, reportTexFile)
}

func (rep *Report) renderPNGsParallel(dash grafana.Dashboard) (err error) {
	var wg sync.WaitGroup
	wg.Add(len(dash.Panels))

	for _, p := range dash.Panels {
		go func(p grafana.Panel) {
			defer wg.Done()
			err = rep.renderPNG(p)
			if err != nil {
				log.Printf("Error creating image for panel: %v", err)
				return
			}
		}(p)
	}

	wg.Wait()
	return
}

func (rep *Report) renderPNG(p grafana.Panel) error {
	body, err := rep.gClient.GetPanelPng(p, rep.dashName, rep.time)
	if err != nil {
		return fmt.Errorf("error getting panel: %v", err)
	}
	defer body.Close()

	err = os.MkdirAll(rep.imgDirPath(), 0777)
	if err != nil {
		return fmt.Errorf("error creating img directory:%v", err)
	}
	imgFileName := fmt.Sprintf("image%d.png", p.Id)
	file, err := os.Create(filepath.Join(rep.imgDirPath(), imgFileName))
	if err != nil {
		return fmt.Errorf("error creating image file:%v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("error copying body to file:%v", err)
	}
	return nil
}

func (rep *Report) generateTeXFile(dash grafana.Dashboard) (err error) {
	type templData struct {
		grafana.Dashboard
		grafana.TimeRange
		grafana.Client
	}

	err = os.MkdirAll(rep.tmpDir, 0777)
	if err != nil {
		return
	}
	file, err := os.Create(rep.texPath())
	if err != nil {
		return
	}
	defer file.Close()

	tmpl, err := template.New("report").Delims("[[", "]]").Parse(rep.texTemplate)
	if err != nil {
		err = fmt.Errorf("Error parsing template '%s': %q", rep.texTemplate, err)
		return
	}
	data := templData{dash, rep.time, rep.gClient}
	err = tmpl.Execute(file, data)
	return
}

func (rep *Report) runLaTeX() (pdf *os.File, err error) {
	cmdPre := exec.Command("pdflatex", "-halt-on-error", "-draftmode", reportTexFile)
	cmdPre.Dir = rep.tmpDir
	outBytesPre, errPre := cmdPre.CombinedOutput()
	log.Println("Calling LaTeX - preprocessing")
	if errPre != nil {
		err = fmt.Errorf("Error calling LaTeX: %q. Latex preprocessing failed with output: %s ", errPre, string(outBytesPre))
		return
	}
	cmd := exec.Command("pdflatex", "-halt-on-error", reportTexFile)
	cmd.Dir = rep.tmpDir
	outBytes, err := cmd.CombinedOutput()
	log.Println("Calling LaTeX and building PDF")
	if err != nil {
		err = fmt.Errorf("Error calling LaTeX: %q. Latex failed with output: %s ", err, string(outBytes))
		return
	}
	pdf, err = os.Open(rep.pdfPath())
	return
}
