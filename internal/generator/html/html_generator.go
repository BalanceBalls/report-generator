package htmlgenerator

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/BalanceBalls/report-generator/internal/report"
)

type HtmlGenerator struct {
	reportsDir string
	tmplName   string
	saveToDisk bool
}

//go:embed *.tmpl
var tpls embed.FS

const pathToBin = "./bin/"

func New(reportsDir string, tmplName string, saveToDisk bool) *HtmlGenerator {
	return &HtmlGenerator{
		reportsDir: reportsDir,
		tmplName:   tmplName,
		saveToDisk: saveToDisk,
	}
}

func (g *HtmlGenerator) Generate(data report.Report) (report.Result, error) {
	tmpl, err := template.ParseFS(tpls, g.tmplName)
	if err != nil {
		return report.Result{}, err
	}
	var reportData bytes.Buffer
	if err := tmpl.ExecuteTemplate(&reportData, g.tmplName, data.Rows); err != nil {
		return report.Result{}, err
	}

	reportName := fmt.Sprintf("%d_%d.html", data.UserId, time.Now().UnixMilli())
	reportBytes := reportData.Bytes()

	if g.saveToDisk {
		path := createReportPath(data.Id, g.reportsDir, reportName)
		if err = createDirIfNotExist(g.reportsDir); err != nil {
			return report.Result{}, err
		}

		file, err := createFileIfNotExist(path)
		if err != nil {
			return report.Result{}, err
		}

		_, err = file.Write(reportBytes)

		if err != nil {
			return report.Result{}, err
		}

		if err = file.Close(); err != nil {
			return report.Result{}, err
		}
	}

	return report.Result{
		Name: reportName, 
		Data: reportData.Bytes(),
	}, nil
}

func createDirIfNotExist(reportsDir string) error {
	pathToDir := filepath.Join(pathToBin, reportsDir)
	if _, err := os.Stat(pathToDir); errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(pathToDir, fs.ModePerm)
	}

	return nil
}

func createFileIfNotExist(path string) (*os.File, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		var f *os.File

		if f, err = os.Create(path); err != nil {
			return nil, err
		}

		if err := f.Chmod(fs.ModePerm); err != nil {
			return nil, err
		}

		return f, nil
	}

	return nil, errors.New("file already exists: " + path)
}

func createReportPath(id int64, reportsDir string, fileName string) string {
	path := filepath.Join(pathToBin, reportsDir, fileName)
	return path
}
