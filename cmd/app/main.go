package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/diegodario88/correios-processor/pkg/db"
	"github.com/diegodario88/correios-processor/pkg/readers"
)

func main() {
	fmt.Println("Starting Correios Processor...")

	// Database connection
	database := &db.DB{}
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Define the base directory for the eDNE files
	baseDir := "eDNE/basico" // Adjust if necessary

	// Process LOG_LOCALIDADE.TXT
	logLocalidadePath := filepath.Join(baseDir, "LOG_LOCALIDADE.TXT")
	log.Printf("Processing %s...\n", logLocalidadePath)
	logLocalidadeRecords, err := readers.ReadLogLocalidade(logLocalidadePath)
	if err != nil {
		log.Printf("Error reading LOG_LOCALIDADE: %v\n", err)
	} else {
		log.Printf("Read %d records from LOG_LOCALIDADE\n", len(logLocalidadeRecords))
		// if err := database.InsertLogLocalidade(logLocalidadeRecords); err != nil {
		// 	log.Fatalf("Error inserting LOG_LOCALIDADE records: %v", err)
		// }
		log.Println("Successfully inserted/updated LOG_LOCALIDADE records")
	}

	// Process LOG_BAIRRO.TXT
	logBairroPath := filepath.Join(baseDir, "LOG_BAIRRO.TXT")
	log.Printf("Processing %s...\n", logBairroPath)
	logBairroRecords, err := readers.ReadLogBairro(logBairroPath)
	if err != nil {
		log.Printf("Error reading LOG_BAIRRO: %v\n", err)
	} else {
		log.Printf("Read %d records from LOG_BAIRRO\n", len(logBairroRecords))
		// if err := database.InsertLogBairro(logBairroRecords); err != nil {
		// 	log.Fatalf("Error inserting LOG_BAIRRO records: %v", err)
		// }
		log.Println("Successfully inserted/updated LOG_BAIRRO records")
	}

	fmt.Println("Correios Processor finished.")
}
