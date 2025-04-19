package readers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"github.com/diegodario88/correios-processor/pkg/models"
)

func ReadLogLocalidade(filePath string) ([]models.LogLocalidade, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '@'     // Set the delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var records []models.LogLocalidade
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}

		// Ensure the row has enough columns
		if len(row) < 6 {
			continue // Skip rows with insufficient data
		}

		record := models.LogLocalidade{}

		// Assign values to struct fields, handling optional fields and type conversions
		if len(row[0]) > 0 {
			record.LocNu, err = strconv.ParseInt(row[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid LOC_NU: %w", err)
			}
		}
		record.UfeSg = row[1]
		record.LocNo = row[2]
		if len(row) > 3 && len(row[3]) > 0 {
			record.Cep = &row[3]
		}
		record.LocInSit = row[4]
		record.LocInTipoLoc = row[5]
		if len(row) > 6 && len(row[6]) > 0 {
			locNuSub, err := strconv.ParseInt(row[6], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid LOC_NU_SUB: %w", err)
			}
			record.LocNuSub = &locNuSub
		}
		if len(row) > 7 && len(row[7]) > 0 {
			record.LocNoAbrev = &row[7]
		}
		if len(row) > 8 && len(row[8]) > 0 {
			record.MunNu = &row[8]
		}

		records = append(records, record)
	}

	return records, nil
}

func ReadLogBairro(filePath string) ([]models.LogBairro, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '@'
	reader.FieldsPerRecord = -1

	var records []models.LogBairro
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}

		if len(row) < 4 {
			continue
		}

		record := models.LogBairro{}

		if len(row[0]) > 0 {
			record.BaiNu, err = strconv.ParseInt(row[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid BAI_NU: %w", err)
			}
		}
		record.UfeSg = row[1]
		if len(row[2]) > 0 {
			record.LocNu, err = strconv.ParseInt(row[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid LOC_NU: %w", err)
			}
		}
		record.BaiNo = row[3]
		if len(row) > 4 && len(row[4]) > 0 {
			record.BaiNoAbrev = &row[4]
		}

		records = append(records, record)
	}

	return records, nil
}
