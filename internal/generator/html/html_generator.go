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

	"github.com/BalanceBalls/report-generator/internal/generator"
	"github.com/BalanceBalls/report-generator/internal/storage"
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

func (g *HtmlGenerator) Generate(report storage.Report) (generator.Report, error) {
	fmt.Println("starting report generation")

	tmpl, err := template.ParseFS(tpls, g.tmplName)
	if err != nil {
		return generator.Report{}, fmt.Errorf(
			"failed to parse template file for html report: %w", err)
	}
	var reportData bytes.Buffer
	if err := tmpl.ExecuteTemplate(&reportData, g.tmplName, report.Rows); err != nil {
		return generator.Report{}, fmt.Errorf(
			"failed to generate an html report: %w", err)
	}

	reportBytes := reportData.Bytes()

	if g.saveToDisk {
		path := createReportPath(report.Id, g.reportsDir)
		if err = createDirIfNotExist(g.reportsDir); err != nil {
			return generator.Report{}, fmt.Errorf(
				"failed to create reports folder: %w", err)
		}

		file, err := createFileIfNotExist(path)
		if err != nil {
			return generator.Report{}, fmt.Errorf(
				"failed to create html file for report: %w", err)
		}

		_, err = file.Write(reportBytes)

		if err != nil {
			return generator.Report{}, fmt.Errorf(
				"failed to save report into file: %w", err)
		}

		if err = file.Close(); err != nil {
			return generator.Report{}, fmt.Errorf(
				"failed to close html file: %w", err)
		}
	}

	return generator.Report{
		Name: fmt.Sprint(report.UserId) + "_" + fmt.Sprint(time.Now().UnixMilli()) + ".html",
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
