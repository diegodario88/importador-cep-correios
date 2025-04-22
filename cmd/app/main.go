package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	immu "github.com/diegodario88/importador-cep-correios/pkg/constants"
	"github.com/diegodario88/importador-cep-correios/pkg/db"
	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	work "github.com/diegodario88/importador-cep-correios/pkg/workers"
)

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	var lineCount int64
	basePath := filepath.Join(utils.GetCWD(), "eDNE", "basico")
	database := &db.DB{}
	ctx := context.Background()
	counterChan := make(chan work.Counter)
	progress := mpb.New(mpb.WithWidth(64))

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Disconnect()

	if err := database.CreateCorreiosTables(); err != nil {
		log.Fatal(err)
	}

	bar := progress.New(int64(immu.SIXTEEN_TASKS),
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.BarFillerOnComplete(""),
		mpb.PrependDecorators(
			decor.Name("Importação base eDNE Correios:"),
			decor.OnComplete(
				decor.Spinner(nil, decor.WCSyncSpace), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)

	run := func(fileName string, execute work.Processes) {
		defer wg.Done()
		defer bar.Increment()

		tools := work.JobTools{
			Ctx:         ctx,
			Database:    database,
			BasePath:    basePath,
			CounterChan: counterChan,
		}

		execute(fileName, tools)
	}

	wg.Add(immu.SIXTEEN_TASKS)
	go run("ECT_PAIS.TXT", work.Single)
	go run("LOG_FAIXA_UF.TXT", work.Single)
	go run("LOG_LOCALIDADE.TXT", work.Single)
	go run("LOG_VAR_LOC.TXT", work.Single)
	go run("LOG_FAIXA_LOCALIDADE.TXT", work.Single)
	go run("LOG_BAIRRO.TXT", work.Single)
	go run("LOG_VAR_BAI.TXT", work.Single)
	go run("LOG_FAIXA_BAIRRO.TXT", work.Single)
	go run("LOG_CPC.TXT", work.Single)
	go run("LOG_FAIXA_CPC.TXT", work.Single)
	go run("LOG_LOGRADOURO_*.TXT", work.Multiple)
	go run("LOG_VAR_LOG.TXT", work.Single)
	go run("LOG_NUM_SEC.TXT", work.Single)
	go run("LOG_GRANDE_USUARIO.TXT", work.Single)
	go run("LOG_UNID_OPER.TXT", work.Single)
	go run("LOG_FAIXA_UOP.TXT", work.Single)

	go func() {
		wg.Wait()
		close(counterChan)
	}()

	for result := range counterChan {
		if result.Error != nil {
			log.Fatalf("Erro no processamento: %v", result.Error)
		}

		lineCount += int64(result.Increment)
	}

	progress.Wait()
	fmt.Println("\nRelatório final:")

	totalRecords, _ := database.GetTotalRecords()
	totalCeps, _ := database.GetTotalCEPs()

	fmt.Printf("Registros totais: %s\n", utils.FormatNumber(totalRecords))
	fmt.Printf("Total de CEPs: %s\n", utils.FormatNumber(totalCeps))
	fmt.Printf("Total de linhas: %s\n", utils.FormatNumber(int(lineCount)))
	fmt.Printf("Tempo total: %s\n", time.Since(start).Round(time.Millisecond))
	log.Println("------------------------------------------------------")
}
