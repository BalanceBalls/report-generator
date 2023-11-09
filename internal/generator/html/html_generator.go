package htmlgenerator

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BalanceBalls/report-generator/internal/generator"
	"github.com/BalanceBalls/report-generator/internal/storage"
)

type HtmlGenerator struct {
	reportsDir string
	tmplName   string
}

//go:embed *.tmpl
var tpls embed.FS

const pathToBin = "./bin/"

func New(reportsDir string, tmplName string) *HtmlGenerator {
	return &HtmlGenerator{
		reportsDir: reportsDir,
		tmplName:   tmplName,
	}
}

func (g *HtmlGenerator) Generate(report storage.Report) (generator.GeneratedReport, error) {
	fmt.Println("starting report generation")

	tmpl, err := template.ParseFS(tpls, g.tmplName)
	if err != nil {
		return generator.GeneratedReport{}, fmt.Errorf(
			"failed to parse template file for html report: %w", err)
	}

	path := createReportPath(report.Id, g.reportsDir)
	if err = createDirIfNotExist(g.reportsDir); err != nil {
		return generator.GeneratedReport{}, fmt.Errorf(
			"failed to create reports folder: %w", err)
	}

	file, err := createFileIfNotExist(path)
	if err != nil {
		return generator.GeneratedReport{}, fmt.Errorf(
			"failed to create html file for report: %w", err)
	}

	// TODO: get bytes array and return 
	if err = tmpl.ExecuteTemplate(file, g.tmplName, report.Rows); err != nil {
		return generator.GeneratedReport{}, fmt.Errorf(
			"failed to generate an html report: %w", err)
	}

	if err = file.Close(); err != nil {
		return generator.GeneratedReport{}, fmt.Errorf(
			"failed to close html file: %w", err)
	}
	
	fmt.Println("report generated")

	return generator.GeneratedReport{}, nil
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

		if f,err = os.Create(path); err != nil {
			return nil, err
		}
		
		if err := f.Chmod(fs.ModePerm); err != nil {
			return nil, err
		}
		
		return f, nil
	}

	return nil, errors.New("file already exists: " + path)
}

func createReportPath(id int, reportsDir string) string {
	fileName := "report-" + strconv.Itoa(id) + ".html"
	path := filepath.Join(pathToBin, reportsDir, fileName)
	return path
}
