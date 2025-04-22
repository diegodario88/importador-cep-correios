package workers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	immu "github.com/diegodario88/importador-cep-correios/pkg/constants"
	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	"golang.org/x/text/encoding/charmap"
)

func Single(fileName string, tools JobTools) {
	filePath := filepath.Join(tools.BasePath, fileName)
	counter := Counter{
		Increment: 1,
		Error:     nil,
	}

	_, err := os.Stat(filePath)
	if err != nil {
		counter.Error = fmt.Errorf("arquivo %s não encontrado: %w", filePath, err)
		tools.CounterChan <- counter
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		counter.Error = fmt.Errorf("erro ao abrir arquivo %s: %w", filePath, err)
		tools.CounterChan <- counter
		return
	}
	defer file.Close()

	decoder := charmap.ISO8859_1.NewDecoder()
	reader := decoder.Reader(file)

	if _, err := file.Seek(0, 0); err != nil {
		counter.Error = fmt.Errorf("erro ao resetar leitura do arquivo: %w", err)
		tools.CounterChan <- counter
		return
	}

	scanner := bufio.NewScanner(reader)
	const batchSize = immu.ONE_THOUSAND_BATCH_SIZE
	var batch [][]any

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "@")
		row := make([]any, len(fields))

		for i := range fields {
			row[i] = utils.HandleEmpty(fields[i], fileName)
		}

		batch = append(batch, row)
		if len(batch) >= batchSize {
			if err := tools.Database.BulkInsertFile(fileName, batch); err != nil {
				counter.Error = err
				tools.CounterChan <- counter
				return
			}
			batch = batch[:0]
		}

		tools.CounterChan <- counter
	}

	if err := scanner.Err(); err != nil {
		counter.Error = fmt.Errorf("erro ao escanear arquivo: %w", err)
		tools.CounterChan <- counter
		return
	}

	if len(batch) > 0 {
		if err := tools.Database.BulkInsertFile(fileName, batch); err != nil {
			counter.Error = err
			tools.CounterChan <- counter
			return
		}
	}
}
