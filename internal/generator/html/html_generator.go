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

func New(reportsDir string, tmplName string, saveTooDisk bool) *HtmlGenerator {
	return &HtmlGenerator{
		reportsDir: reportsDir,
		tmplName:   tmplName,
	}
}

func (g *HtmlGenerator) Generate(data report.Report) (report.Result, error) {
	fmt.Println("starting report generation")

	tmpl, err := template.ParseFS(tpls, g.tmplName)
	if err != nil {
		return report.Result{}, fmt.Errorf(
			"failed to parse template file for html report: %w", err)
	}
	var reportData bytes.Buffer
	if err := tmpl.ExecuteTemplate(&reportData, g.tmplName, data.Rows); err != nil {
		return report.Result{}, fmt.Errorf(
			"failed to generate an html report: %w", err)
	}

	reportBytes := reportData.Bytes()

	if g.saveToDisk {
		path := createReportPath(data.Id, g.reportsDir)
		if err = createDirIfNotExist(g.reportsDir); err != nil {
			return report.Result{}, fmt.Errorf(
				"failed to create reports folder: %w", err)
		}

		file, err := createFileIfNotExist(path)
		if err != nil {
			return report.Result{}, fmt.Errorf(
				"failed to create html file for report: %w", err)
		}

		_, err = file.Write(reportBytes)

		if err != nil {
			return report.Result{}, fmt.Errorf(
				"failed to save report into file: %w", err)
		}

		if err = file.Close(); err != nil {
			return report.Result{}, fmt.Errorf(
				"failed to close html file: %w", err)
		}
	}

	return report.Result{
		Name: fmt.Sprint(data.UserId) + "_" + fmt.Sprint(time.Now().UnixMilli()) + ".html",
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

func createReportPath(id int64, reportsDir string) string {
	fileName := "report-" + fmt.Sprint(id) + ".html"
	path := filepath.Join(pathToBin, reportsDir, fileName)
	return path
}
