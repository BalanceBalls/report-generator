package builder

import "github.com/BalanceBalls/report-generator/internal/storage"

type Builder interface {
	Build() (*storage.Report, error)
}

