package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/diegodario88/correios-processor/pkg/models"
)

type DB struct {
	Conn *sql.DB
}

func (db *DB) Connect() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRESQL_HOST"),
		os.Getenv("POSTGRESQL_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	db.Conn = conn
	log.Println("Successfully connected to the database")
	return nil
}

func (db *DB) InsertLogLocalidade(records []models.LogLocalidade) error {
	for _, record := range records {
		_, err := db.Conn.Exec(`
			INSERT INTO correios.log_localidade (loc_nu, ufe_sg, loc_no, cep, loc_in_sit, loc_in_tipo_loc, loc_nu_sub, loc_no_abrev, mun_nu)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (loc_nu) DO UPDATE SET
				ufe_sg = EXCLUDED.ufe_sg,
				loc_no = EXCLUDED.loc_no,
				cep = EXCLUDED.cep,
				loc_in_sit = EXCLUDED.loc_in_sit,
				loc_in_tipo_loc = EXCLUDED.loc_in_tipo_loc,
				loc_nu_sub = EXCLUDED.loc_nu_sub,
				loc_no_abrev = EXCLUDED.loc_no_abrev,
				mun_nu = EXCLUDED.mun_nu;
		`,
			record.LocNu, record.UfeSg, record.LocNo, record.Cep, record.LocInSit, record.LocInTipoLoc, record.LocNuSub, record.LocNoAbrev, record.MunNu)
		if err != nil {
			return fmt.Errorf("error inserting log_localidade record: %w", err)
		}
	}
	return nil
}

func (db *DB) InsertLogBairro(records []models.LogBairro) error {
	for _, record := range records {
		_, err := db.Conn.Exec(`
			INSERT INTO correios.log_bairro (bai_nu, ufe_sg, loc_nu, bai_no, bai_no_abrev)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (bai_nu) DO UPDATE SET
				ufe_sg = EXCLUDED.ufe_sg,
				loc_nu = EXCLUDED.loc_nu,
				bai_no = EXCLUDED.bai_no,
				bai_no_abrev = EXCLUDED.bai_no_abrev;
		`,
			record.BaiNu, record.UfeSg, record.LocNu, record.BaiNo, record.BaiNoAbrev)
		if err != nil {
			return fmt.Errorf("error inserting log_bairro record: %w", err)
		}
	}
	return nil
}

func (db *DB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close()
	}
	return nil
}
