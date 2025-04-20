package workers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/diegodario88/importador-cep-correios/pkg/utils"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/text/encoding/charmap"
)

func ProcessLogLogradouro(tools JobTools) error {
	var wg sync.WaitGroup
	filePattern := "LOG_LOGRADOURO_*.TXT"
	matches, err := filepath.Glob(filepath.Join(tools.BasePath, filePattern))
	if err != nil {
		return fmt.Errorf("erro ao buscar arquivos de logradouro: %w", err)
	}

	ufs := len(matches)
	if ufs == 0 {
		return fmt.Errorf("nenhum arquivo de logradouro encontrado com o padrão %s", filePattern)
	}

	errCh := make(chan error, ufs)
	wg.Add(ufs)

	for _, filePath := range matches {
		fileName := filepath.Base(filePath)

		lineCount, err := utils.CountLines(filePath)
		if err != nil {
			return fmt.Errorf("erro ao contar linhas: %w", err)
		}

		ufBar := tools.Progress.AddBar(
			int64(lineCount),
			mpb.PrependDecorators(
				decor.Name(fileName, decor.WC{W: len(fileName) + 1, C: decor.DindentRight}),
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "importado",
				),
			),
		)

		go processLogradouroFile(
			tools,
			filePath,
			fileName,
			ufBar,
			errCh,
			&wg,
		)
	}

	wg.Wait()
	close(errCh)

	var errMsgs []string
	for err := range errCh {
		errMsgs = append(errMsgs, err.Error())
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("errors creating tables: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

func processLogradouroFile(
	tools JobTools,
	filePath, fileName string,
	bar *mpb.Bar,
	c chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		c <- fmt.Errorf("erro ao abrir arquivo %s: %w", fileName, err)
		return
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
			if err := tools.Database.BulkInsertLogLogradouro(batch); err != nil {
				c <- err
				return
			}
			batch = batch[:0]
		}

		bar.Increment()
	}

	if err := scanner.Err(); err != nil {
		c <- fmt.Errorf("erro ao escanear arquivo: %w", err)
		return
	}

	if len(batch) > 0 {
		if err := tools.Database.BulkInsertLogLogradouro(batch); err != nil {
			c <- err
			return
		}
	}
}
