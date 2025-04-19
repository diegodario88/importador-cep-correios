package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/diegodario88/correios-processor/pkg/models"
)

type DB struct {
	Conn *sql.DB
}

func (db *DB) Connect() error {
	err := godotenv.Load("/app/.env")
	if err != nil {
		log.Fatal(err)
	}

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

func (db *DB) Version() (string, error) {
	row := db.Conn.QueryRow("SELECT version()")
	var fullVersion string
	if err := row.Scan(&fullVersion); err != nil {
		return "", fmt.Errorf("erro ao obter versão do PostgreSQL: %w", err)
	}

	if idx := strings.Index(fullVersion, "("); idx != -1 {
		return strings.TrimSpace(fullVersion[:idx]), nil
	}
	return fullVersion, nil
}

func (db *DB) CreateCorreiosSchema() error {
	_, err := db.Conn.Exec("CREATE SCHEMA IF NOT EXISTS correios;")
	if err != nil {
		return fmt.Errorf("erro ao criar schema correios: %w", err)
	}
	return nil
}

func (db *DB) GetTotalRecords() (int, error) {
	query := `
	SELECT sum((xpath('/row/cnt/text()', xml_count))[1]::TEXT::int ) AS total_records
	FROM (
		SELECT
			table_name,
			table_schema,
			query_to_xml(format('select count(*) as cnt from %I.%I', table_schema, table_name), FALSE, TRUE, '') AS xml_count
		FROM information_schema.tables
		WHERE table_schema = 'correios'
	) t;`

	var total int
	if err := db.Conn.QueryRow(query).Scan(&total); err != nil {
		return 0, fmt.Errorf("erro ao obter total de registros: %w", err)
	}
	return total, nil
}

func (db *DB) GetTotalCEPs() (int, error) {
	query := `
	SELECT CAST(COUNT(*) AS INTEGER) as total_ceps
	FROM (
		SELECT cep FROM correios.log_localidade
		UNION ALL
		SELECT cep FROM correios.log_logradouro
		UNION ALL
		SELECT cep FROM correios.log_grande_usuario
		UNION ALL
		SELECT cep FROM correios.log_unid_oper
		UNION ALL
		SELECT cep FROM correios.log_cpc
	) as all_ceps;`

	var total int
	if err := db.Conn.QueryRow(query).Scan(&total); err != nil {
		return 0, fmt.Errorf("erro ao obter total de CEPs: %w", err)
	}
	return total, nil
}

func (db *DB) InsertLogLocalidade(records []models.LogLocalidade) error {
	for _, record := range records {
		_, err := db.Conn.Exec(
			`
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
			record.LocNu,
			record.UfeSg,
			record.LocNo,
			record.Cep,
			record.LocInSit,
			record.LocInTipoLoc,
			record.LocNuSub,
			record.LocNoAbrev,
			record.MunNu,
		)
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
