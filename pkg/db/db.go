package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	immu "github.com/diegodario88/importador-cep-correios/pkg/constants"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type DB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

func (db *DB) Connect() error {
	db.ctx = context.Background()

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRESQL_HOST"),
		os.Getenv("POSTGRESQL_PORT"),
		os.Getenv("POSTGRES_DB"))

	pool, err := pgxpool.New(db.ctx, connStr)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	err = pool.Ping(db.ctx)
	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	db.Pool = pool
	version, err := db.Version()
	if err != nil {
		return fmt.Errorf("error seeking for database version: %w", err)
	}

	log.Println("------------------------------------------------------")
	log.Println("Successfully connected to the database")
	log.Println(version)
	return nil
}

func (db *DB) Disconnect() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Disconnected from database")
		log.Println("------------------------------------------------------")
	}
}

func (db *DB) Version() (string, error) {
	row := db.Pool.QueryRow(db.ctx, "SELECT version()")
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
	_, err := db.Pool.Exec(db.ctx, "CREATE SCHEMA IF NOT EXISTS correios;")
	if err != nil {
		return fmt.Errorf("erro ao criar schema correios: %w", err)
	}
	return nil
}

func (db *DB) CreateCorreiosTables() error {
	var wg sync.WaitGroup
	errChan := make(chan error, immu.SIXTEEN_TASKS)

	createTable := func(name string, createFn func() error) {
		defer wg.Done()
		if err := createFn(); err != nil {
			log.Printf("Error creating %s: %v", name, err)
			errChan <- fmt.Errorf("error creating %s: %w", name, err)
		}
	}

	if err := db.CreateCorreiosSchema(); err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	wg.Add(immu.SIXTEEN_TASKS)

	go createTable("ect_pais", db.createTableECTPais)
	go createTable("log_faixa_uf", db.createTableLogFaixaUF)
	go createTable("log_localidade", db.createTableLogLocalidade)
	go createTable("log_var_loc", db.createTableLogVarLoc)
	go createTable("log_faixa_localidade", db.createTableLogFaixaLocalidade)
	go createTable("log_bairro", db.createTableLogBairro)
	go createTable("log_var_bai", db.createTableLogVarBai)
	go createTable("log_faixa_bairro", db.createTableLogFaixaBairro)
	go createTable("log_cpc", db.createTableLogCPC)
	go createTable("log_faixa_cpc", db.createTableLogFaixaCPC)
	go createTable("log_logradouro", db.createTableLogLogradouro)
	go createTable("log_var_log", db.createTableLogVarLog)
	go createTable("log_num_sec", db.createTableLogNumSec)
	go createTable("log_grande_usuario", db.createTableLogGrandeUsuario)
	go createTable("log_unid_oper", db.createTableLogUnidOper)
	go createTable("log_faixa_uop", db.createTableLogFaixaUOP)

	wg.Wait()
	close(errChan)

	var errMsgs []string
	for err := range errChan {
		errMsgs = append(errMsgs, err.Error())
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("errors creating tables: %s", strings.Join(errMsgs, "; "))
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
	if err := db.Pool.QueryRow(db.ctx, query).Scan(&total); err != nil {
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
	if err := db.Pool.QueryRow(db.ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("erro ao obter total de CEPs: %w", err)
	}
	return total, nil
}

func (db *DB) BulkInsertFile(fileName string, rows [][]any) error {
	if strings.HasPrefix(fileName, "LOG_LOGRADOURO_") && strings.HasSuffix(fileName, ".TXT") {
		return db.bulkInsertLogLogradouro(rows)
	}

	switch fileName {
	case "ECT_PAIS.TXT":
		return db.bulkInsertECTPais(rows)
	case "LOG_FAIXA_UF.TXT":
		return db.bulkInsertLogFaixaUF(rows)
	case "LOG_LOCALIDADE.TXT":
		return db.bulkInsertLogLocalidade(rows)
	case "LOG_VAR_LOC.TXT":
		return db.bulkInsertLogVarLoc(rows)
	case "LOG_FAIXA_LOCALIDADE.TXT":
		return db.bulkInsertLogFaixaLocalidade(rows)
	case "LOG_BAIRRO.TXT":
		return db.bulkInsertLogBairro(rows)
	case "LOG_VAR_BAI.TXT":
		return db.bulkInsertLogVarBai(rows)
	case "LOG_FAIXA_BAIRRO.TXT":
		return db.bulkInsertLogFaixaBairro(rows)
	case "LOG_CPC.TXT":
		return db.bulkInsertLogCPC(rows)
	case "LOG_FAIXA_CPC.TXT":
		return db.bulkInsertLogFaixaCPC(rows)
	case "LOG_VAR_LOG.TXT":
		return db.bulkInsertLogVarLog(rows)
	case "LOG_NUM_SEC.TXT":
		return db.bulkInsertLogNumSec(rows)
	case "LOG_GRANDE_USUARIO.TXT":
		return db.bulkInsertLogGrandeUsuario(rows)
	case "LOG_UNID_OPER.TXT":
		return db.bulkInsertLogUnidOper(rows)
	case "LOG_FAIXA_UOP.TXT":
		return db.bulkInsertLogFaixaUOP(rows)
	default:
		return fmt.Errorf("unknown file name: %s", fileName)
	}
}

func (db *DB) createTableLogFaixaUF() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_faixa_uf(
		ufe_sg char(2) NOT NULL,
		ufe_cep_ini char(8) NOT NULL,
		ufe_cep_fim char(8) NOT NULL,
		PRIMARY KEY (ufe_sg, ufe_cep_ini)
	);
	COMMENT on column correios.log_faixa_uf.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_faixa_uf.ufe_cep_ini is 'CEP inicial da UF';
	COMMENT on column correios.log_faixa_uf.ufe_cep_fim is 'CEP final da UF';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_faixa_uf table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogFaixaUF(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_faixa_uf"},
		[]string{"ufe_sg", "ufe_cep_ini", "ufe_cep_fim"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_faixa_uf: %w", err)
	}
	return nil
}

func (db *DB) createTableLogLocalidade() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_localidade(
		loc_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_no varchar(72) NOT NULL,
		cep char(8) NULL,
		loc_in_sit char(1) NOT NULL,
		loc_in_tipo_loc char(1) NOT NULL,
		loc_nu_sub numeric NULL,
		loc_no_abrev varchar(36) NULL,
		mun_nu char(7) NULL,
		PRIMARY KEY (loc_nu)
	);
	COMMENT on column correios.log_localidade.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_localidade.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_localidade.loc_no is 'nome da localidade';
	COMMENT on column correios.log_localidade.cep is 'CEP da localidade (para localidade não codificada, ou seja loc_in_sit = 0)';
	COMMENT on column correios.log_localidade.loc_in_sit is '0 = Localidade não codificada em nível de Logradouro,1 = Localidade codificada em nível de Logradouro, 2 = Distrito ou Povoado inserido na codificação em nível de Logradouro, 3 = Localidade em fase de codificação em nível de Logradouro.';
	COMMENT on column correios.log_localidade.loc_in_tipo_loc is 'tipo de localidade: D – Distrito,M – Município,P – Povoado.';
	COMMENT on column correios.log_localidade.loc_nu_sub is 'chave da localidade de subordinação';
	COMMENT on column correios.log_localidade.loc_no_abrev is 'abreviatura do nome da localidade';
	COMMENT on column correios.log_localidade.mun_nu is 'Código do município IBGE';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_localidade table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogLocalidade(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_localidade"},
		[]string{
			"loc_nu", "ufe_sg", "loc_no", "cep",
			"loc_in_sit", "loc_in_tipo_loc", "loc_nu_sub",
			"loc_no_abrev", "mun_nu",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_localidade: %w", err)
	}
	return nil
}

func (db *DB) createTableLogVarLoc() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_var_loc(
		loc_nu numeric NOT NULL,
		val_nu numeric NOT NULL,
		val_tx varchar(72) NOT NULL,
		PRIMARY KEY (loc_nu, val_nu)
	);
	COMMENT on column correios.log_var_loc.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_var_loc.val_nu is 'ordem da localidade';
	COMMENT on column correios.log_var_loc.val_tx is 'Denominação';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_var_loc table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogVarLoc(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_var_loc"},
		[]string{"loc_nu", "val_nu", "val_tx"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_var_loc: %w", err)
	}
	return nil
}

func (db *DB) createTableLogFaixaLocalidade() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_faixa_localidade(
		loc_nu numeric NOT NULL,
		loc_cep_ini char(8) NOT NULL,
		loc_cep_fim char(8) NOT NULL,
		loc_tipo_faixa char(1) NOT NULL,
		PRIMARY KEY (loc_nu, loc_cep_ini, loc_tipo_faixa)
	);
	COMMENT on column correios.log_faixa_localidade.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_faixa_localidade.loc_cep_ini is 'CEP inicial da localidade';
	COMMENT on column correios.log_faixa_localidade.loc_cep_fim is 'CEP final da localidade';
	COMMENT on column correios.log_faixa_localidade.loc_tipo_faixa is 'tipo de Faixa de CEP:T –Total do Município C – Exclusiva da  Sede Urbana';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_faixa_localidade table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogFaixaLocalidade(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_faixa_localidade"},
		[]string{"loc_nu", "loc_cep_ini", "loc_cep_fim", "loc_tipo_faixa"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_faixa_localidade: %w", err)
	}
	return nil
}

func (db *DB) createTableLogBairro() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_bairro(
		bai_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_nu char(8) NOT NULL,
		bai_no varchar(72) NOT NULL,
		bai_no_abrev varchar(36) NULL,
		PRIMARY KEY (bai_nu)
	);
	COMMENT on column correios.log_bairro.bai_nu is 'chave do bairro';
	COMMENT on column correios.log_bairro.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_bairro.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_bairro.bai_no is 'nome do bairro';
	COMMENT on column correios.log_bairro.bai_no_abrev is 'abreviatura do nome do bairro';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_bairro table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogBairro(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_bairro"},
		[]string{"bai_nu", "ufe_sg", "loc_nu", "bai_no", "bai_no_abrev"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_bairro: %w", err)
	}
	return nil
}

func (db *DB) createTableLogVarBai() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_var_bai(
		bai_nu numeric NOT NULL,
		vdb_nu char(2) NOT NULL,
		vdb_tx varchar(72) NOT NULL,
		PRIMARY KEY (bai_nu, vdb_nu)
	);
	COMMENT on column correios.log_var_bai.bai_nu is 'chave do bairro';
	COMMENT on column correios.log_var_bai.vdb_nu is 'ordem da denominação';
	COMMENT on column correios.log_var_bai.vdb_tx is 'Denominação';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_var_bai table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogVarBai(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_var_bai"},
		[]string{"bai_nu", "vdb_nu", "vdb_tx"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_var_bai: %w", err)
	}
	return nil
}

func (db *DB) createTableLogFaixaBairro() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_faixa_bairro(
		bai_nu numeric NOT NULL,
		fcb_cep_ini char(8) NOT NULL,
		fcb_cep_fim char(8) NOT NULL,
		PRIMARY KEY (bai_nu, fcb_cep_ini)
	);
	COMMENT on column correios.log_faixa_bairro.bai_nu is 'chave do bairro';
	COMMENT on column correios.log_faixa_bairro.fcb_cep_ini is 'CEP inicial do bairro';
	COMMENT on column correios.log_faixa_bairro.fcb_cep_fim is 'CEP final do bairro';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_faixa_bairro table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogFaixaBairro(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_faixa_bairro"},
		[]string{"bai_nu", "fcb_cep_ini", "fcb_cep_fim"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_faixa_bairro: %w", err)
	}
	return nil
}

func (db *DB) createTableLogCPC() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_cpc(
		cpc_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_nu numeric NOT NULL,
		cpc_no varchar(72) NOT NULL,
		cpc_endereco varchar(100) NOT NULL,
		cep char(8) NOT NULL,
		PRIMARY KEY (cpc_nu)
	);
	COMMENT on column correios.log_cpc.cpc_nu is 'chave da caixa postal comunitária';
	COMMENT on column correios.log_cpc.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_cpc.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_cpc.cpc_no is 'nome da CPC';
	COMMENT on column correios.log_cpc.cpc_endereco is 'endereço da CPC';
	COMMENT on column correios.log_cpc.cep is 'CEP da CPC';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_cpc table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogCPC(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_cpc"},
		[]string{"cpc_nu", "ufe_sg", "loc_nu", "cpc_no", "cpc_endereco", "cep"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_cpc: %w", err)
	}
	return nil
}

func (db *DB) createTableLogFaixaCPC() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_faixa_cpc(
		cpc_nu numeric NOT NULL,
		cpc_inicial varchar(6) NOT NULL,
		cpc_final varchar(6) NOT NULL,
		PRIMARY KEY (cpc_nu, cpc_inicial)
	);
	COMMENT on column correios.log_faixa_cpc.cpc_nu is 'chave da caixa postal comunitária';
	COMMENT on column correios.log_faixa_cpc.cpc_inicial is 'número inicial da caixa postal comunitária';
	COMMENT on column correios.log_faixa_cpc.cpc_final is 'número final da caixa postal comunitária';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_faixa_cpc table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogFaixaCPC(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_faixa_cpc"},
		[]string{"cpc_nu", "cpc_inicial", "cpc_final"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_faixa_cpc: %w", err)
	}
	return nil
}

func (db *DB) createTableLogLogradouro() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_logradouro(
		log_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_nu numeric NOT NULL,
		bai_nu_ini numeric NOT NULL,
		bai_nu_fim numeric NULL,
		log_no varchar(100) NOT NULL,
		log_complemento varchar(100) NULL,
		cep char(8) NOT NULL,
		tlo_tx varchar(100) NOT NULL,
		log_sta_tlo char(1) NULL,
		log_no_abrev varchar(100) NULL,
		PRIMARY KEY (log_nu)
	);
	COMMENT on column correios.log_logradouro.log_nu is 'chave do logradouro';
	COMMENT on column correios.log_logradouro.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_logradouro.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_logradouro.bai_nu_ini is 'chave do bairro inicial do logradouro';
	COMMENT on column correios.log_logradouro.bai_nu_fim is 'chave do bairro final do logradouro';
	COMMENT on column correios.log_logradouro.log_no is 'nome do logradouro';
	COMMENT on column correios.log_logradouro.log_complemento is 'complemento do logradouro';
	COMMENT on column correios.log_logradouro.cep is 'CEP do logradouro';
	COMMENT on column correios.log_logradouro.tlo_tx is 'tipo de logradouro';
	COMMENT on column correios.log_logradouro.log_sta_tlo is 'indicador de utilização do tipo de logradouro (S ou N)';
	COMMENT on column correios.log_logradouro.log_no_abrev is 'abreviatura do nome do logradouro';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_logradouro table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogLogradouro(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_logradouro"},
		[]string{"log_nu", "ufe_sg", "loc_nu", "bai_nu_ini", "bai_nu_fim",
			"log_no", "log_complemento", "cep", "tlo_tx", "log_sta_tlo", "log_no_abrev"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_logradouro: %w", err)
	}
	return nil
}

func (db *DB) createTableLogVarLog() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_var_log(
		log_nu numeric NOT NULL,
		vlo_nu numeric NOT NULL,
		tlo_tx varchar(36) NOT NULL,
		vlo_tx varchar(150) NOT NULL,
		PRIMARY KEY (log_nu, vlo_nu)
	);
	COMMENT on column correios.log_var_log.log_nu is 'chave do logradouro';
	COMMENT on column correios.log_var_log.vlo_nu is 'ordem da denominação';
	COMMENT on column correios.log_var_log.tlo_tx is 'tipo de logradouro da variação';
	COMMENT on column correios.log_var_log.vlo_tx is 'nome da variação do logradouro';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_var_log table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogVarLog(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_var_log"},
		[]string{"log_nu", "vlo_nu", "tlo_tx", "vlo_tx"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_var_log: %w", err)
	}
	return nil
}

func (db *DB) createTableLogNumSec() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_num_sec(
		log_nu numeric NOT NULL,
		sec_nu_ini varchar(10) NOT NULL,
		sec_nu_fim varchar(10) NOT NULL,
		sec_in_lado char(1) NOT NULL,
		PRIMARY KEY (log_nu)
	);
	COMMENT on column correios.log_num_sec.log_nu is 'chave do logradouro';
	COMMENT on column correios.log_num_sec.sec_nu_ini is 'número inicial do seccionamento';
	COMMENT on column correios.log_num_sec.sec_nu_fim is 'número final do seccionamento';
	COMMENT on column correios.log_num_sec.sec_in_lado is 'Indica a paridade/lado do seccionamento A – ambos,P – par,I – ímpar,D – direito eE – esquerdo.';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_num_sec table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogNumSec(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_num_sec"},
		[]string{"log_nu", "sec_nu_ini", "sec_nu_fim", "sec_in_lado"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_num_sec: %w", err)
	}
	return nil
}

func (db *DB) createTableLogGrandeUsuario() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_grande_usuario(
		gru_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_nu numeric NOT NULL,
		bai_nu numeric NOT NULL,
		log_nu numeric NULL,
		gru_no varchar(255) NOT NULL,
		gru_endereco varchar(255) NOT NULL,
		cep char(8) NOT NULL,
		gru_no_abrev varchar(255) NULL,
		PRIMARY KEY (gru_nu)
	);
	COMMENT on column correios.log_grande_usuario.gru_nu is 'chave do grande usuário';
	COMMENT on column correios.log_grande_usuario.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_grande_usuario.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_grande_usuario.bai_nu is 'chave do bairro';
	COMMENT on column correios.log_grande_usuario.log_nu is 'chave do logradouro';
	COMMENT on column correios.log_grande_usuario.gru_no is 'nome do grande usuário';
	COMMENT on column correios.log_grande_usuario.gru_endereco is 'endereço do grande usuário';
	COMMENT on column correios.log_grande_usuario.cep is 'CEP do grande usuário';
	COMMENT on column correios.log_grande_usuario.gru_no_abrev is 'abreviatura do nome do grande usuário';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_grande_usuario table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogGrandeUsuario(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_grande_usuario"},
		[]string{"gru_nu", "ufe_sg", "loc_nu", "bai_nu", "log_nu",
			"gru_no", "gru_endereco", "cep", "gru_no_abrev"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_grande_usuario: %w", err)
	}
	return nil
}

func (db *DB) createTableLogUnidOper() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_unid_oper(
		uop_nu numeric NOT NULL,
		ufe_sg char(2) NOT NULL,
		loc_nu numeric NOT NULL,
		bai_nu numeric NOT NULL,
		log_nu numeric NULL,
		uop_no varchar(100) NOT NULL,
		uop_endereco varchar(100) NOT NULL,
		cep char(8) NOT NULL,
		uop_in_cp char(1) NOT NULL,
		uop_no_abrev varchar(100) NULL,
		PRIMARY KEY (uop_nu)
	);
	COMMENT on column correios.log_unid_oper.uop_nu is 'chave da UOP';
	COMMENT on column correios.log_unid_oper.ufe_sg is 'sigla da UF';
	COMMENT on column correios.log_unid_oper.loc_nu is 'chave da localidade';
	COMMENT on column correios.log_unid_oper.bai_nu is 'chave do bairro';
	COMMENT on column correios.log_unid_oper.log_nu is 'chave do logradouro';
	COMMENT on column correios.log_unid_oper.uop_no is 'nome da UOP';
	COMMENT on column correios.log_unid_oper.uop_endereco is 'endereço da UOP';
	COMMENT on column correios.log_unid_oper.cep is 'CEP da UOP';
	COMMENT on column correios.log_unid_oper.uop_in_cp is 'indicador de caixa postal (S ou N)';
	COMMENT on column correios.log_unid_oper.uop_no_abrev is 'abreviatura do nome da unid. operacional';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_unid_oper table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogUnidOper(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_unid_oper"},
		[]string{"uop_nu", "ufe_sg", "loc_nu", "bai_nu", "log_nu",
			"uop_no", "uop_endereco", "cep", "uop_in_cp", "uop_no_abrev"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_unid_oper: %w", err)
	}
	return nil
}

func (db *DB) createTableLogFaixaUOP() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.log_faixa_uop(
		uop_nu numeric NOT NULL,
		fnc_inicial numeric NOT NULL,
		fnc_final numeric NOT NULL,
		PRIMARY KEY (uop_nu, fnc_inicial)
	);
	COMMENT on column correios.log_faixa_uop.uop_nu is 'chave da UOP';
	COMMENT on column correios.log_faixa_uop.fnc_inicial is 'número inicial da caixa postal';
	COMMENT on column correios.log_faixa_uop.fnc_final is 'número final da caixa postal';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating log_faixa_uop table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertLogFaixaUOP(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "log_faixa_uop"},
		[]string{"uop_nu", "fnc_inicial", "fnc_final"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into log_faixa_uop: %w", err)
	}
	return nil
}

func (db *DB) createTableECTPais() error {
	query := `
	CREATE TABLE IF NOT EXISTS correios.ect_pais(
		pai_sg char(2) NOT NULL,
		pai_sg_alternativa char(3) NOT NULL,
		pai_no_portugues varchar(100) NOT NULL,
		pai_no_ingles varchar(100) NOT NULL,
		pai_no_frances varchar(100) NOT NULL,
		pai_abreviatura varchar(100) NOT NULL,
		PRIMARY KEY (pai_sg)
	);
	COMMENT on column correios.ect_pais.pai_sg is 'Sigla do País';
	COMMENT on column correios.ect_pais.pai_sg_alternativa is 'Sigla alternativa';
	`

	_, err := db.Pool.Exec(db.ctx, query)
	if err != nil {
		return fmt.Errorf("error creating ect_pais table: %w", err)
	}
	return nil
}

func (db *DB) bulkInsertECTPais(rows [][]any) error {
	_, err := db.Pool.CopyFrom(
		db.ctx,
		pgx.Identifier{"correios", "ect_pais"},
		[]string{"pai_sg", "pai_sg_alternativa", "pai_no_portugues",
			"pai_no_ingles", "pai_no_frances", "pai_abreviatura"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("error bulk inserting into ect_pais: %w", err)
	}
	return nil
}
