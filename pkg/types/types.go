package types

import (
	"context"
	"time"
)

type Storage interface {
	Connect() error
	Disconnect()
	Version() (string, error)
	CreateCorreiosSchema() error
	CreateCorreiosSql() error
	GetTotalRecords() (int, error)
	GetTotalCEPs() (int, error)
	BulkInsertFile(fileName string, rows [][]any) error
	GetCep(cep string) (CepResponse, error)
	InsertImportacaoRelatorio(input ImportacaoRelatorio) error
}

type Counter struct {
	Increment int
	Error     error
}

type JobTools struct {
	Ctx         context.Context
	Database    Storage
	BasePath    string
	CounterChan chan<- Counter
}

type Processes func(string, JobTools)

type CepResponse struct {
	uf          string
	localidade  string
	cep         string
	ibge        string
	bairro      *string
	complemento *string
	logradouro  *string
}

type ImportacaoRelatorio struct {
	TotalRegistros int
	TotalCeps      int
	VersaoEDNE     string
	Duracao        time.Duration
	Observacoes    string
}
