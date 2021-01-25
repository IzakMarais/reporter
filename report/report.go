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

	"github.com/iTransformers/reporter/grafana"
	"github.com/pborman/uuid"
)

// Report groups functions related to genrating the report.
// After reading and closing the pdf returned by Generate(), call Clean() to delete the pdf file as well the temporary build files
type Report interface {
	Generate() (pdf io.ReadCloser, err error)
	Title() string
	Clean()
}

type report struct {
	gClient     grafana.Client
	time        grafana.TimeRange
	texTemplate string
	dashName    string
	tmpDir      string
	dashTitle   string
}

const (
	imgDir        = "images"
	reportTexFile = "report.tex"
	reportPdf     = "report.pdf"
)

// New creates a new Report.
// texTemplate is the content of a LaTex template file. If empty, a default tex template is used.
func New(g grafana.Client, dashName string, time grafana.TimeRange, texTemplate string, gridLayout bool) Report {
	return new(g, dashName, time, texTemplate, gridLayout)
}

func new(g grafana.Client, dashName string, time grafana.TimeRange, texTemplate string, gridLayout bool) *report {
	if texTemplate == "" {
		if gridLayout {
			texTemplate = defaultGridTemplate
		} else {
			texTemplate = defaultTemplate
		}

	}
	tmpDir := filepath.Join("tmp", uuid.New())
	return &report{g, time, texTemplate, dashName, tmpDir, ""}
}

// Generate returns the report.pdf file.  After reading this file it should be Closed()
// After closing the file, call report.Clean() to delete the file as well the temporary build files
func (rep *report) Generate() (pdf io.ReadCloser, err error) {
	dash, err := rep.gClient.GetDashboard(rep.dashName)
	if err != nil {
		err = fmt.Errorf("error fetching dashboard %v: %v", rep.dashName, err)
		return
	}
	rep.dashTitle = dash.Title

	err = rep.renderPNGsParallel(dash)
	if err != nil {
		err = fmt.Errorf("error rendering PNGs in parralel for dash %+v: %v", dash, err)
		return
	}
	err = rep.generateTeXFile(dash)
	if err != nil {
		err = fmt.Errorf("error generating TeX file for dash %+v: %v", dash, err)
		return
	}
	pdf, err = rep.runLaTeX()
	return
}

// Title returns the dashboard title parsed from the dashboard definition
func (rep *report) Title() string {
	//lazy fetch if Title() is called before Generate()
	if rep.dashTitle == "" {
		dash, err := rep.gClient.GetDashboard(rep.dashName)
		if err != nil {
			return ""
		}
		rep.dashTitle = dash.Title
	}
	return rep.dashTitle
}

// Clean deletes the temporary directory used during report generation
func (rep *report) Clean() {
	err := os.RemoveAll(rep.tmpDir)
	if err != nil {
		log.Println("Error cleaning up tmp dir:", err)
	}
}

func (rep *report) imgDirPath() string {
	return filepath.Join(rep.tmpDir, imgDir)
}

func (rep *report) pdfPath() string {
	return filepath.Join(rep.tmpDir, reportPdf)
}

func (rep *report) texPath() string {
	return filepath.Join(rep.tmpDir, reportTexFile)
}

func (rep *report) renderPNGsParallel(dash grafana.Dashboard) error {
	//buffer all panels on a channel
	panels := make(chan grafana.Panel, len(dash.Panels))
	for _, p := range dash.Panels {
		panels <- p
	}
	close(panels)

	//fetch images in parrallel form Grafana sever.
	//limit concurrency using a worker pool to avoid overwhelming grafana
	//for dashboards with many panels.
	var wg sync.WaitGroup
	workers := 5
	wg.Add(workers)
	errs := make(chan error, len(dash.Panels)) //routines can return errors on a channel
	for i := 0; i < workers; i++ {
		go func(panels <-chan grafana.Panel, errs chan<- error) {
			defer wg.Done()
			for p := range panels {
				err := rep.renderPNG(p)
				if err != nil {
					log.Printf("Error creating image for panel: %v", err)
					errs <- err
				}
			}
		}(panels, errs)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (rep *report) renderPNG(p grafana.Panel) error {
	body, err := rep.gClient.GetPanelPng(p, rep.dashName, rep.time)
	if err != nil {
		return fmt.Errorf("error getting panel %+v: %v", p, err)
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

func (rep *report) generateTeXFile(dash grafana.Dashboard) error {
	type templData struct {
		grafana.Dashboard
		grafana.TimeRange
		grafana.Client
	}

	err := os.MkdirAll(rep.tmpDir, 0777)
	if err != nil {
		return fmt.Errorf("error creating temporary directory at %v: %v", rep.tmpDir, err)
	}
	file, err := os.Create(rep.texPath())
	if err != nil {
		return fmt.Errorf("error creating tex file at %v : %v", rep.texPath(), err)
	}
	defer file.Close()

	tmpl, err := template.New("report").Delims("[[", "]]").Parse(rep.texTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template '%s': %v", rep.texTemplate, err)
	}
	data := templData{dash, rep.time, rep.gClient}
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing tex template:%v", err)
	}
	return nil
}

func (rep *report) runLaTeX() (pdf *os.File, err error) {
	cmdPre := exec.Command("pdflatex", "-halt-on-error", "-draftmode", reportTexFile)
	cmdPre.Dir = rep.tmpDir
	outBytesPre, errPre := cmdPre.CombinedOutput()
	log.Println("Calling LaTeX - preprocessing")
	if errPre != nil {
		err = fmt.Errorf("error calling LaTeX preprocessing: %q. Latex preprocessing failed with output: %s ", errPre, string(outBytesPre))
		return
	}
	cmd := exec.Command("pdflatex", "-halt-on-error", reportTexFile)
	cmd.Dir = rep.tmpDir
	outBytes, err := cmd.CombinedOutput()
	log.Println("Calling LaTeX and building PDF")
	if err != nil {
		err = fmt.Errorf("error calling LaTeX: %q. Latex failed with output: %s ", err, string(outBytes))
		return
	}
	pdf, err = os.Open(rep.pdfPath())
	return
}
