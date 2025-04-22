package workers

import (
	"context"

	"github.com/diegodario88/importador-cep-correios/pkg/db"
)

type Counter struct {
	Increment int
	Error     error
}

type JobTools struct {
	Ctx         context.Context
	Database    *db.DB
	BasePath    string
	CounterChan chan<- Counter
}

type Processes func(string, JobTools)
