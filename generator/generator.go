package generator

import "github.com/BalanceBalls/report-generator/storage"

type Generator interface {
	Generate(report storage.Report) (GeneratedReport, error)
}

type GeneratedReport struct {
	Data []byte
}
