package generator

import "github.com/BalanceBalls/report-generator/internal/storage"

type Generator interface {
	Generate(report storage.Report) (Report, error)
}

type Report struct {
	Name string
	Data []byte
}
