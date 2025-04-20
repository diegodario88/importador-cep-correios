package workers

import (
	"context"

	"github.com/diegodario88/importador-cep-correios/pkg/db"
	"github.com/vbauerster/mpb/v8"
)

type Job struct {
	Ctx      context.Context
	Database *db.DB
	BasePath string
	Progress *mpb.Progress
}

type Processes func(Job) error
