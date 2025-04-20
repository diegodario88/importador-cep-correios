package workers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/text/encoding/charmap"
)

func ProcessLogFaixaLocalidade(job Job) error {
	fileName := "LOG_FAIXA_LOCALIDADE.TXT"
	filePath := filepath.Join(job.BasePath, fileName)

	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("arquivo %s não encontrado: %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := charmap.ISO8859_1.NewDecoder()
	reader := decoder.Reader(file)

	lineCount, err := utils.CountLines(filePath)
	if err != nil {
		return fmt.Errorf("erro ao contar linhas: %w", err)
	}

	bar := job.Progress.AddBar(int64(lineCount),
		mpb.PrependDecorators(
			decor.Name(fileName, decor.WC{W: len(fileName) + 1, C: decor.DindentRight}),
			decor.Percentage(decor.WC{W: 6}),
		),
	)

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("erro ao resetar leitura do arquivo: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	const batchSize = 1000
	var batch [][]any

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "@")
		row := make([]any, len(fields))

		for i := range fields {
			if fields[i] == "" {
				row[i] = nil
			} else {
				row[i] = strings.TrimSpace(fields[i])
			}
		}

		batch = append(batch, row)
		if len(batch) >= batchSize {
			if err := job.Database.BulkInsertLogFaixaLocalidade(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}

		bar.Increment()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("erro ao escanear arquivo: %w", err)
	}

	if len(batch) > 0 {
		if err := job.Database.BulkInsertLogFaixaLocalidade(batch); err != nil {
			return err
		}
	}

	return nil
}
