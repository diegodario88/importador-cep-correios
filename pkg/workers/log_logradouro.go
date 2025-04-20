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

func ProcessLogLogradouro(job Job) error {
	filePattern := "LOG_LOGRADOURO_*.TXT"
	matches, err := filepath.Glob(filepath.Join(job.BasePath, filePattern))
	if err != nil {
		return fmt.Errorf("erro ao buscar arquivos de logradouro: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("nenhum arquivo de logradouro encontrado com o padrão %s", filePattern)
	}

	var totalLines int64
	for _, filePath := range matches {
		lines, err := utils.CountLines(filePath)
		if err != nil {
			return fmt.Errorf(
				"erro ao contar linhas do arquivo %s: %w",
				filepath.Base(filePath),
				err,
			)
		}
		totalLines += int64(lines)
	}

	bar := job.Progress.AddBar(totalLines,
		mpb.PrependDecorators(
			decor.Name("LOG_LOGRADOURO", decor.WC{W: 15, C: decor.DindentRight}),
			decor.Percentage(decor.WC{W: 6}),
		),
	)

	for _, filePath := range matches {
		fileName := filepath.Base(filePath)
		if err := processLogradouroFile(job, filePath, fileName, bar); err != nil {
			return fmt.Errorf("erro ao processar %s: %w", fileName, err)
		}
	}

	return nil
}

func processLogradouroFile(job Job, filePath, fileName string, bar *mpb.Bar) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", fileName, err)
	}
	defer file.Close()

	decoder := charmap.ISO8859_1.NewDecoder()
	reader := decoder.Reader(file)

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
			if err := job.Database.BulkInsertLogLogradouro(batch); err != nil {
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
		if err := job.Database.BulkInsertLogLogradouro(batch); err != nil {
			return err
		}
	}

	return nil
}
