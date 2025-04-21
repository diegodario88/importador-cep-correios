package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"

	immu "github.com/diegodario88/importador-cep-correios/pkg/constants"
	"github.com/diegodario88/importador-cep-correios/pkg/db"
	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	work "github.com/diegodario88/importador-cep-correios/pkg/workers"
)

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	basePath := filepath.Join(utils.GetCWD(), "eDNE", "basico")
	database := &db.DB{}
	ctx := context.Background()
	errCh := make(chan error, immu.SIXTEEN_TASKS)
	progress := mpb.New(
		mpb.WithWidth(64),
		mpb.WithWaitGroup(&wg),
		mpb.WithOutput(os.Stderr),
		mpb.WithAutoRefresh(),
	)

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Disconnect()

	if err := database.CreateCorreiosTables(); err != nil {
		log.Fatal(err)
	}

	run := func(execute work.Processes) {
		defer wg.Done()

		tools := work.JobTools{
			Ctx:      ctx,
			Database: database,
			BasePath: basePath,
			Progress: progress,
		}

		if err := execute(tools); err != nil {
			errCh <- err
		}

	}

	wg.Add(immu.SIXTEEN_TASKS)

	go run(work.ProcessECTPais)
	go run(work.ProcessLogFaixaUF)
	go run(work.ProcessLogLocalidade)
	go run(work.ProcessLogVarLoc)
	go run(work.ProcessLogFaixaLocalidade)
	go run(work.ProcessLogBairro)
	go run(work.ProcessLogVarBai)
	go run(work.ProcessLogFaixaBairro)
	go run(work.ProcessLogCPC)
	go run(work.ProcessLogFaixaCPC)
	go run(work.ProcessLogLogradouro)
	go run(work.ProcessLogVarLog)
	go run(work.ProcessLogNumSec)
	go run(work.ProcessLogGrandeUsuario)
	go run(work.ProcessLogUnidOper)
	go run(work.ProcessLogFaixaUOP)

	wg.Wait()
	close(errCh)

	for err := range errCh {
		log.Fatalf("Erro no processamento: %v", err)
	}

	progress.Wait()
	fmt.Println("\nRelatório final:")

	totalRecords, _ := database.GetTotalRecords()
	totalCeps, _ := database.GetTotalCEPs()

	fmt.Printf("Registros totais: %s\n", utils.FormatNumber(totalRecords))
	fmt.Printf("Total de CEPs: %s\n", utils.FormatNumber(totalCeps))
	fmt.Printf("Tempo total: %s\n", time.Since(start).Round(time.Millisecond))
}
