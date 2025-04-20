package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vbauerster/mpb/v8"

	"github.com/diegodario88/importador-cep-correios/pkg/db"
	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	"github.com/diegodario88/importador-cep-correios/pkg/workers"
)

func main() {
	start := time.Now()
	basePath := filepath.Join(utils.GetCWD(), "eDNE", "basico")
	progress := mpb.New(mpb.WithWidth(60), mpb.WithOutput(os.Stdout))
	database := &db.DB{}
	ctx := context.Background()

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Disconnect()

	if err := database.CreateCorreiosSchema(); err != nil {
		log.Fatal(err)
	}

	if err := database.CreateCorreiosTables(); err != nil {
		log.Fatal(err)
	}

	executors := []workers.Processes{
		workers.ProcessECTPais,
		workers.ProcessLogFaixaUF,
		workers.ProcessLogLocalidade,
		workers.ProcessLogVarLoc,
		workers.ProcessLogFaixaLocalidade,
		workers.ProcessLogBairro,
		workers.ProcessLogVarBai,
		workers.ProcessLogFaixaBairro,
		workers.ProcessLogCPC,
		workers.ProcessLogFaixaCPC,
		workers.ProcessLogLogradouro,
		workers.ProcessLogVarLog,
		workers.ProcessLogNumSec,
		workers.ProcessLogGrandeUsuario,
		workers.ProcessLogUnidOper,
		workers.ProcessLogFaixaUOP,
	}

	errCh := make(chan error, len(executors))
	done := make(chan struct{})

	for _, worker := range executors {
		go func(execute workers.Processes) {
			job := workers.Job{
				Ctx:      ctx,
				Database: database,
				BasePath: basePath,
				Progress: progress,
			}

			if err := execute(job); err != nil {
				errCh <- err
			}
			done <- struct{}{}
		}(worker)
	}

	for range executors {
		select {
		case <-done:
		case err := <-errCh:
			log.Fatalf("Erro no processamento: %v", err)
		}
	}

	progress.Wait()

	fmt.Println("\nRelatório final:")

	totalRecords, _ := database.GetTotalRecords()
	totalCeps, _ := database.GetTotalCEPs()

	fmt.Printf("Registros totais: %s\n", utils.FormatNumber(totalRecords))
	fmt.Printf("Total de CEPs: %s\n", utils.FormatNumber(totalCeps))
	fmt.Printf("Tempo total: %s\n", time.Since(start).Round(time.Millisecond))
}
