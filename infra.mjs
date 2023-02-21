import postgres from "pg";

export class InfrastructureService {
  DATABASE_CLIENT;

  constructor() {
    const { Client } = postgres;
    this.DATABASE_CLIENT = new Client({
      host: process.env.POSTGRESQL_HOST,
      port: process.env.POSTGRESQL_PORT,
      user: process.env.POSTGRES_USER,
      password: process.env.POSTGRES_PASSWORD,
      database: process.env.POSTGRES_DB,
      connectionTimeoutMillis: 5000,
    });
  }

  async connectToDatabase() {
    await this.DATABASE_CLIENT.connect();
    const version = await this.version();
    console.log(
      `Connected to ${this.DATABASE_CLIENT.host}@${this.DATABASE_CLIENT.database} ${version}`
    );
  }

  async disconnectToDatabase() {
    await this.DATABASE_CLIENT.end();
    console.warn("-".repeat(30));
    console.warn(`Disconnected to database ${this.DATABASE_CLIENT.host}`);
  }

  /**
   * @returns {Promise<string>}
   */
  async version() {
    const query = {
      name: "select-version",
      text: "SELECT version()",
    };

    const versionQueryResult = await this.DATABASE_CLIENT.query(query);
    const version = versionQueryResult.rows[0].version.split("(").at(0);
    return version;
  }

  async createCorreiosSchema() {
    const query = {
      name: "create-schema",
      text: "CREATE SCHEMA IF NOT EXISTS correios;",
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @returns {Promise<number>}
   */
  async getTotalRecords() {
    const query = {
      name: "get-total-records",
      text: `SELECT sum((xpath('/row/cnt/text()', xml_count))[1]::TEXT::int ) AS total_records
              FROM
              (
              SELECT
                table_name,
                table_schema,
                query_to_xml(format('select count(*) as cnt from %I.%I', table_schema, table_name), FALSE, TRUE, '') AS xml_count
              FROM
                information_schema.tables
              WHERE
                table_schema = 'correios'
              ) t`,
    };

    const result = await this.DATABASE_CLIENT.query(query);
    return result.rows[0].total_records;
  }

  /**
   * @returns {Promise<number>}
   */
  async getTotalCEPS() {
    const query = {
      name: "get-total-ceps",
      text: `SELECT CAST(COUNT(*) AS INTEGER) as total_ceps
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
              ) as all_ceps;`,
    };

    const result = await this.DATABASE_CLIENT.query(query);
    return result.rows[0].total_ceps;
  }

  async CREATE_TABLE_LOG_FAIXA_UF() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_FAIXA_UF(
        UFE_SG char(2) NOT NULL,
        UFE_CEP_INI char(8) NOT NULL,
        UFE_CEP_FIM char(8) NOT NULL,
        PRIMARY KEY (UFE_SG, UFE_CEP_INI)
      );
      COMMENT on column correios.LOG_FAIXA_UF.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_FAIXA_UF.UFE_CEP_INI is 'CEP inicial da UF';
      COMMENT on column correios.LOG_FAIXA_UF.UFE_CEP_FIM is 'CEP final da UF';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_FAIXA_UF(values) {
    const query = {
      text: `INSERT INTO correios.LOG_FAIXA_UF(
        UFE_SG, 
        UFE_CEP_INI, 
        UFE_CEP_FIM
        ) 
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (UFE_SG, UFE_CEP_INI) DO NOTHING;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_LOCALIDADE() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_LOCALIDADE(
        LOC_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NO varchar(72) NOT NULL,
        CEP char(8) NULL,
        LOC_IN_SIT char(1) NOT NULL,
        LOC_IN_TIPO_LOC char(1) NOT NULL,
        LOC_NU_SUB numeric  NULL,
        LOC_NO_ABREV varchar(36)  NULL,
        MUN_NU char(7)  NULL,
        PRIMARY KEY (LOC_NU)
      );
      COMMENT on column correios.LOG_LOCALIDADE.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_LOCALIDADE.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_LOCALIDADE.LOC_NO is 'nome da localidade';
      COMMENT on column correios.LOG_LOCALIDADE.CEP is 'CEP da localidade (para  localidade  não codificada, ou seja loc_in_sit = 0)';
      COMMENT on column correios.LOG_LOCALIDADE.LOC_IN_SIT is '0 = Localidade não codificada em nível de Logradouro,1 = Localidade codificada em nível de Logradouro, 2 = Distrito ou Povoado inserido na codificação em nível de Logradouro, 3 = Localidade em fase de codificação em nível de Logradouro.';
      COMMENT on column correios.LOG_LOCALIDADE.LOC_IN_TIPO_LOC is 'tipo de localidade: D – Distrito,M – Município,P – Povoado.';
      COMMENT on column correios.LOG_LOCALIDADE.LOC_NU_SUB is 'chave da localidade de subordinação';
      COMMENT on column correios.LOG_LOCALIDADE.LOC_NO_ABREV is 'abreviatura do nome da localidade';
      COMMENT on column correios.LOG_LOCALIDADE.MUN_NU is 'Código do município IBGE';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_LOCALIDADE(values) {
    const query = {
      text: `INSERT INTO correios.LOG_LOCALIDADE(
        LOC_NU,
        UFE_SG,
        LOC_NO,
        CEP,
        LOC_IN_SIT,
        LOC_IN_TIPO_LOC,
        LOC_NU_SUB,
        LOC_NO_ABREV ,
        MUN_NU
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6,
          $7,
          $8,
          $9
          )
          ON CONFLICT (LOC_NU) DO UPDATE SET
            UFE_SG = EXCLUDED.UFE_SG,
            LOC_NO = EXCLUDED.LOC_NO,
            CEP = EXCLUDED.CEP,
            LOC_IN_SIT = EXCLUDED.LOC_IN_SIT,
            LOC_IN_TIPO_LOC = EXCLUDED.LOC_IN_TIPO_LOC,
            LOC_NU_SUB = EXCLUDED.LOC_NU_SUB,
            LOC_NO_ABREV = EXCLUDED.LOC_NO_ABREV,
            MUN_NU = EXCLUDED.MUN_NU
          ;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} LOC_NU
   */
  async DELETE_FROM_LOG_LOCALIDADE(LOC_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_LOCALIDADE WHERE LOC_NU = $1`,
      values: [LOC_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_VAR_LOC() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_VAR_LOC(
        LOC_NU numeric NOT NULL,
        VAL_NU numeric NOT NULL,
        VAL_TX varchar(72) NOT NULL,
        PRIMARY KEY (LOC_NU, VAL_NU)
      );
      COMMENT on column correios.LOG_VAR_LOC.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_VAR_LOC.VAL_NU is 'ordem da localidade';
      COMMENT on column correios.LOG_VAR_LOC.VAL_TX is 'Denominação';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_VAR_LOC(values) {
    const query = {
      text: `INSERT INTO correios.LOG_VAR_LOC(
        LOC_NU,
        VAL_NU,
        VAL_TX
        )
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (LOC_NU, VAL_NU) DO NOTHING;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_FAIXA_LOCALIDADE() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_FAIXA_LOCALIDADE(
        LOC_NU numeric NOT NULL,
        LOC_CEP_INI char(8) NOT NULL,
        LOC_CEP_FIM char(8) NOT NULL,
        LOC_TIPO_FAIXA char(1) NOT NULL,
        PRIMARY KEY (LOC_NU, LOC_CEP_INI, LOC_TIPO_FAIXA)
      );
      COMMENT on column correios.LOG_FAIXA_LOCALIDADE.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_FAIXA_LOCALIDADE.LOC_CEP_INI is 'CEP inicial da localidade';
      COMMENT on column correios.LOG_FAIXA_LOCALIDADE.LOC_CEP_FIM is 'CEP final da localidade';
      COMMENT on column correios.LOG_FAIXA_LOCALIDADE.LOC_TIPO_FAIXA is 'tipo de Faixa de CEP:T –Total do Município C – Exclusiva da  Sede Urbana';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_FAIXA_LOCALIDADE(values) {
    const query = {
      text: `INSERT INTO correios.LOG_FAIXA_LOCALIDADE(
        LOC_NU,
        LOC_CEP_INI,
        LOC_CEP_FIM,
        LOC_TIPO_FAIXA
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4
          )
          ON CONFLICT (LOC_NU, LOC_CEP_INI, LOC_TIPO_FAIXA) DO UPDATE SET
            LOC_TIPO_FAIXA = EXCLUDED.LOC_TIPO_FAIXA;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array} primaries
   */
  async DELETE_FROM_LOG_FAIXA_LOCALIDADE(primaries) {
    const query = {
      text: `DELETE FROM correios.LOG_FAIXA_LOCALIDADE
             WHERE LOC_NU = $1 AND LOC_CEP_INI = $2 AND LOC_TIPO_FAIXA = $3`,
      values: primaries,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_BAIRRO() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_BAIRRO(
        BAI_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NU char(8) NOT NULL,
        BAI_NO varchar(72) NOT NULL,
        BAI_NO_ABREV varchar(36) NULL,
        PRIMARY KEY (BAI_NU)
      );
      COMMENT on column correios.LOG_BAIRRO.BAI_NU is 'chave do bairro';
      COMMENT on column correios.LOG_BAIRRO.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_BAIRRO.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_BAIRRO.BAI_NO is 'nome do bairro';
      COMMENT on column correios.LOG_BAIRRO.BAI_NO_ABREV is 'abreviatura do nome do bairro';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_BAIRRO(values) {
    const query = {
      text: `INSERT INTO correios.LOG_BAIRRO(
        BAI_NU,
        UFE_SG,
        LOC_NU,
        BAI_NO,
        BAI_NO_ABREV
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5
          )
          ON CONFLICT (BAI_NU) DO UPDATE SET
            UFE_SG = excluded.UFE_SG,
            LOC_NU = excluded.LOC_NU,
            BAI_NO = excluded.BAI_NO,
            BAI_NO_ABREV = excluded.BAI_NO_ABREV;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} BAI_NU
   */
  async DELETE_FROM_LOG_BAIRRO(BAI_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_BAIRRO WHERE BAI_NU = $1`,
      values: [BAI_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_VAR_BAI() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_VAR_BAI(
        BAI_NU numeric NOT NULL,
        VDB_NU char(2) NOT NULL,
        VDB_TX varchar(72) NOT NULL,
        PRIMARY KEY (BAI_NU, VDB_NU)
      );
      COMMENT on column correios.LOG_VAR_BAI.BAI_NU is 'chave do bairro';
      COMMENT on column correios.LOG_VAR_BAI.VDB_NU is 'ordem da denominação';
      COMMENT on column correios.LOG_VAR_BAI.VDB_TX is 'Denominação';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_VAR_BAI(values) {
    const query = {
      text: `INSERT INTO correios.LOG_VAR_BAI(
        BAI_NU,
        VDB_NU,
        VDB_TX
        )
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (BAI_NU, VDB_NU) DO NOTHING;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_FAIXA_BAIRRO() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_FAIXA_BAIRRO(
        BAI_NU numeric NOT NULL,
        FCB_CEP_INI char(8) NOT NULL,
        FCB_CEP_FIM char(8) NOT NULL,
        PRIMARY KEY (BAI_NU, FCB_CEP_INI)
      );
      COMMENT on column correios.LOG_FAIXA_BAIRRO.BAI_NU is 'chave do bairro';
      COMMENT on column correios.LOG_FAIXA_BAIRRO.FCB_CEP_INI is 'CEP inicial do bairro';
      COMMENT on column correios.LOG_FAIXA_BAIRRO.FCB_CEP_FIM is 'CEP final do bairro';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_FAIXA_BAIRRO(values) {
    const query = {
      text: `INSERT INTO correios.LOG_FAIXA_BAIRRO(
        BAI_NU,
        FCB_CEP_INI,
        FCB_CEP_FIM
        )
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (BAI_NU, FCB_CEP_INI) DO UPDATE SET
            FCB_CEP_FIM = EXCLUDED.FCB_CEP_FIM;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array} primaries
   */
  async DELETE_FROM_LOG_FAIXA_BAIRRO(primaries) {
    const query = {
      text: `DELETE FROM correios.LOG_FAIXA_BAIRRO WHERE BAI_NU = $1 AND FCB_CEP_INI = $2`,
      values: primaries,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_CPC() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_CPC(
        CPC_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NU numeric NOT NULL,
        CPC_NO varchar(72) NOT NULL,
        CPC_ENDERECO varchar(100) NOT NULL,
        CEP char(8) NOT NULL,
        PRIMARY KEY (CPC_NU)
      );
      COMMENT on column correios.LOG_CPC.CPC_NU is 'chave da caixa postal comunitária';
      COMMENT on column correios.LOG_CPC.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_CPC.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_CPC.CPC_NO is 'nome da CPC';
      COMMENT on column correios.LOG_CPC.CPC_ENDERECO is 'endereço da CPC';
      COMMENT on column correios.LOG_CPC.CEP is 'CEP da CPC';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_CPC(values) {
    const query = {
      text: `INSERT INTO correios.LOG_CPC(
        CPC_NU,
        UFE_SG,
        LOC_NU,
        CPC_NO,
        CPC_ENDERECO,
        CEP
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6
          )
          ON CONFLICT (CPC_NU) DO UPDATE SET
            UFE_SG = EXCLUDED.UFE_SG,
            LOC_NU = EXCLUDED.LOC_NU,
            CPC_NO = EXCLUDED.CPC_NO,
            CPC_ENDERECO = EXCLUDED.CPC_ENDERECO,
            CEP = EXCLUDED.CEP
          ;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} CPC_NU
   */
  async DELETE_FROM_LOG_CPC(CPC_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_CPC WHERE CPC_NU = $1`,
      values: [CPC_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_FAIXA_CPC() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_FAIXA_CPC(
        CPC_NU numeric NOT NULL,
        CPC_INICIAL varchar(6) NOT NULL,
        CPC_FINAL varchar(6) NOT NULL,
        PRIMARY KEY (CPC_NU, CPC_INICIAL)
      );
      COMMENT on column correios.LOG_FAIXA_CPC.CPC_NU is 'chave da caixa postal comunitária';
      COMMENT on column correios.LOG_FAIXA_CPC.CPC_INICIAL is 'número inicial da caixa postal comunitária';
      COMMENT on column correios.LOG_FAIXA_CPC.CPC_FINAL is 'número final da caixa postal comunitária';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_FAIXA_CPC(values) {
    const query = {
      text: `INSERT INTO correios.LOG_FAIXA_CPC(
        CPC_NU,
        CPC_INICIAL,
        CPC_FINAL
        )
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (CPC_NU, CPC_INICIAL) DO UPDATE SET
            CPC_FINAL = EXCLUDED.CPC_FINAL;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array} primaries
   */
  async DELETE_FROM_LOG_FAIXA_CPC(primaries) {
    const query = {
      text: `DELETE FROM correios.LOG_FAIXA_CPC WHERE CPC_NU = $1 AND CPC_INICIAL = $2`,
      values: primaries,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_LOGRADOURO() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_LOGRADOURO(
        LOG_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NU numeric NOT NULL,
        BAI_NU_INI numeric NOT NULL,
        BAI_NU_FIM numeric NULL,
        LOG_NO varchar(100) NOT NULL,
        LOG_COMPLEMENTO varchar(100) NULL,
        CEP char(8) NOT NULL,
        TLO_TX varchar(100) NOT NULL,
        LOG_STA_TLO char(1) NULL,
        LOG_NO_ABREV varchar(100) NULL,
        PRIMARY KEY (LOG_NU)
      );
      COMMENT on column correios.LOG_LOGRADOURO.LOG_NU is 'chave do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_LOGRADOURO.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_LOGRADOURO.BAI_NU_INI is 'chave do bairro inicial do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.BAI_NU_FIM is 'chave do bairro final do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.LOG_NO is 'nome do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.LOG_COMPLEMENTO is 'complemento do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.CEP is 'CEP do logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.TLO_TX is 'tipo de logradouro';
      COMMENT on column correios.LOG_LOGRADOURO.LOG_STA_TLO is 'indicador de utilização do tipo de logradouro (S ou N)';
      COMMENT on column correios.LOG_LOGRADOURO.LOG_NO_ABREV is 'abreviatura do nome do logradouro';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_LOGRADOURO(values) {
    const query = {
      text: `INSERT INTO correios.LOG_LOGRADOURO(
        LOG_NU,
        UFE_SG,
        LOC_NU,
        BAI_NU_INI,
        BAI_NU_FIM,
        LOG_NO,
        LOG_COMPLEMENTO,
        CEP,
        TLO_TX,
        LOG_STA_TLO,
        LOG_NO_ABREV
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6,
          $7,
          $8,
          $9,
          $10,
          $11
          )
          ON CONFLICT (LOG_NU) DO UPDATE SET
            UFE_SG = EXCLUDED.UFE_SG,
            LOC_NU = EXCLUDED.LOC_NU,
            BAI_NU_INI = EXCLUDED.BAI_NU_INI,
            BAI_NU_FIM = EXCLUDED.BAI_NU_FIM,
            LOG_NO = EXCLUDED.LOG_NO,
            LOG_COMPLEMENTO = EXCLUDED.LOG_COMPLEMENTO,
            CEP = EXCLUDED.CEP,
            TLO_TX = EXCLUDED.TLO_TX,
            LOG_STA_TLO = EXCLUDED.LOG_STA_TLO,
            LOG_NO_ABREV = EXCLUDED.LOG_NO_ABREV;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} LOG_NU
   */
  async DELETE_FROM_LOG_LOGRADOURO(LOG_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_LOGRADOURO WHERE LOG_NU = $1`,
      values: [LOG_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_VAR_LOG() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_VAR_LOG(
        LOG_NU numeric NOT NULL,
        VLO_NU numeric NOT NULL,
        TLO_TX varchar(36) NOT NULL,
        VLO_TX varchar(150) NOT NULL,
        PRIMARY KEY (LOG_NU, VLO_NU)
      );
      COMMENT on column correios.LOG_VAR_LOG.LOG_NU is 'chave do logradouro';
      COMMENT on column correios.LOG_VAR_LOG.VLO_NU is 'ordem da denominação';
      COMMENT on column correios.LOG_VAR_LOG.TLO_TX is 'tipo de logradouro da variação';
      COMMENT on column correios.LOG_VAR_LOG.VLO_TX is 'nome da variação do logradouro';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_VAR_LOG(values) {
    const query = {
      text: `INSERT INTO correios.LOG_VAR_LOG(
        LOG_NU,
        VLO_NU,
        TLO_TX,
        VLO_TX
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4
          )
          ON CONFLICT (LOG_NU, VLO_NU) DO NOTHING;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_NUM_SEC() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_NUM_SEC(
        LOG_NU numeric NOT NULL,
        SEC_NU_INI varchar(10) NOT NULL,
        SEC_NU_FIM varchar(10) NOT NULL,
        SEC_IN_LADO char(1) NOT NULL,
        PRIMARY KEY (LOG_NU)
      );
      COMMENT on column correios.LOG_NUM_SEC.LOG_NU is 'chave do logradouro';
      COMMENT on column correios.LOG_NUM_SEC.SEC_NU_INI is 'número inicial do seccionamento';
      COMMENT on column correios.LOG_NUM_SEC.SEC_NU_FIM is 'número final do seccionamento';
      COMMENT on column correios.LOG_NUM_SEC.SEC_IN_LADO is 'Indica a paridade/lado do seccionamento A – ambos,P – par,I – ímpar,D – direito eE – esquerdo.';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_NUM_SEC(values) {
    const query = {
      text: `INSERT INTO correios.LOG_NUM_SEC(
        LOG_NU,
        SEC_NU_INI,
        SEC_NU_FIM,
        SEC_IN_LADO
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4
          )
          ON CONFLICT (LOG_NU) DO UPDATE SET
            SEC_NU_INI = EXCLUDED.SEC_NU_INI,
            SEC_NU_FIM = EXCLUDED.SEC_NU_FIM,
            SEC_IN_LADO = EXCLUDED.SEC_IN_LADO;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} LOG_NU
   */
  async DELETE_FROM_LOG_NUM_SEC(LOG_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_NUM_SEC WHERE LOG_NU = $1`,
      values: [LOG_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_GRANDE_USUARIO() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_GRANDE_USUARIO(
        GRU_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NU numeric NOT NULL,
        BAI_NU numeric NOT NULL,
        LOG_NU numeric NULL,
        GRU_NO varchar(255) NOT NULL,
        GRU_ENDERECO varchar(255) NOT NULL,
        CEP char(8) NOT NULL,
        GRU_NO_ABREV varchar(255) NULL,
        PRIMARY KEY (GRU_NU)
      );
      COMMENT on column correios.LOG_GRANDE_USUARIO.GRU_NU is 'chave do grande usuário';
      COMMENT on column correios.LOG_GRANDE_USUARIO.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_GRANDE_USUARIO.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_GRANDE_USUARIO.BAI_NU is 'chave do bairro';
      COMMENT on column correios.LOG_GRANDE_USUARIO.LOG_NU is 'chave do logradouro';
      COMMENT on column correios.LOG_GRANDE_USUARIO.GRU_NO is 'nome do grande usuário';
      COMMENT on column correios.LOG_GRANDE_USUARIO.GRU_ENDERECO is 'endereço do grande usuário';
      COMMENT on column correios.LOG_GRANDE_USUARIO.CEP is 'CEP do grande usuário';
      COMMENT on column correios.LOG_GRANDE_USUARIO.GRU_NO_ABREV is 'abreviatura do nome do grande usuário';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_GRANDE_USUARIO(values) {
    const query = {
      text: `INSERT INTO correios.LOG_GRANDE_USUARIO(
        GRU_NU,
        UFE_SG,
        LOC_NU,
        BAI_NU,
        LOG_NU,
        GRU_NO,
        GRU_ENDERECO,
        CEP,
        GRU_NO_ABREV
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6,
          $7,
          $8,
          $9
          )
          ON CONFLICT (GRU_NU) DO UPDATE SET
           UFE_SG = EXCLUDED.UFE_SG,
           LOC_NU = EXCLUDED.LOC_NU,
           BAI_NU = EXCLUDED.BAI_NU,
           LOG_NU = EXCLUDED.LOG_NU,
           GRU_NO = EXCLUDED.GRU_NO,
           GRU_ENDERECO = EXCLUDED.GRU_ENDERECO,
           CEP = EXCLUDED.CEP,
           GRU_NO_ABREV = EXCLUDED.GRU_NO_ABREV
          ;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} GRU_NU
   */
  async DELETE_FROM_LOG_GRANDE_USUARIO(GRU_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_GRANDE_USUARIO WHERE GRU_NU = $1`,
      values: [GRU_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_UNID_OPER() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_UNID_OPER(
        UOP_NU numeric NOT NULL,
        UFE_SG char(2) NOT NULL,
        LOC_NU numeric NOT NULL,
        BAI_NU numeric NOT NULL,
        LOG_NU numeric NULL,
        UOP_NO varchar(100) NOT NULL,
        UOP_ENDERECO varchar(100) NOT NULL,
        CEP char(8) NOT NULL,
        UOP_IN_CP char(1) NOT NULL,
        UOP_NO_ABREV varchar(100) NULL,
        PRIMARY KEY (UOP_NU)
      );
      COMMENT on column correios.LOG_UNID_OPER.UOP_NU is 'chave da UOP';
      COMMENT on column correios.LOG_UNID_OPER.UFE_SG is 'sigla da UF';
      COMMENT on column correios.LOG_UNID_OPER.LOC_NU is 'chave da localidade';
      COMMENT on column correios.LOG_UNID_OPER.BAI_NU is 'chave do bairro';
      COMMENT on column correios.LOG_UNID_OPER.LOG_NU is 'chave do logradouro';
      COMMENT on column correios.LOG_UNID_OPER.UOP_NO is 'nome da UOP';
      COMMENT on column correios.LOG_UNID_OPER.UOP_ENDERECO is 'endereço da UOP';
      COMMENT on column correios.LOG_UNID_OPER.CEP is 'CEP da UOP';
      COMMENT on column correios.LOG_UNID_OPER.UOP_IN_CP is 'indicador de caixa postal (S ou N)';
      COMMENT on column correios.LOG_UNID_OPER.UOP_NO_ABREV is 'abreviatura do nome da unid. operacional';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_UNID_OPER(values) {
    const query = {
      text: `INSERT INTO correios.LOG_UNID_OPER(
        UOP_NU,
        UFE_SG,
        LOC_NU,
        BAI_NU,
        LOG_NU,
        UOP_NO,
        UOP_ENDERECO,
        CEP,
        UOP_IN_CP,
        UOP_NO_ABREV
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6,
          $7,
          $8,
          $9,
          $10
          )
          ON CONFLICT (UOP_NU) DO UPDATE SET
            UFE_SG = EXCLUDED.UFE_SG,
            LOC_NU = EXCLUDED.LOC_NU,
            BAI_NU = EXCLUDED.BAI_NU,
            LOG_NU = EXCLUDED.LOG_NU,
            UOP_NO = EXCLUDED.UOP_NO,
            UOP_ENDERECO = EXCLUDED.UOP_ENDERECO,
            CEP = EXCLUDED.CEP,
            UOP_IN_CP = EXCLUDED.UOP_IN_CP,
            UOP_NO_ABREV = EXCLUDED.UOP_NO_ABREV;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {number} UOP_NU
   */
  async DELETE_FROM_LOG_UNID_OPER(UOP_NU) {
    const query = {
      text: `DELETE FROM correios.LOG_UNID_OPER WHERE UOP_NU = $1`,
      values: [UOP_NU],
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_LOG_FAIXA_UOP() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.LOG_FAIXA_UOP(
        UOP_NU numeric NOT NULL,
        FNC_INICIAL numeric NOT NULL,
        FNC_FINAL numeric NOT NULL,
        PRIMARY KEY (UOP_NU, FNC_INICIAL)
      );
      COMMENT on column correios.LOG_FAIXA_UOP.UOP_NU is 'chave da UOP';
      COMMENT on column correios.LOG_FAIXA_UOP.FNC_INICIAL is 'número inicial da caixa postal';
      COMMENT on column correios.LOG_FAIXA_UOP.FNC_FINAL is 'número final da caixa postal';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_LOG_FAIXA_UOP(values) {
    const query = {
      text: `INSERT INTO correios.LOG_FAIXA_UOP(
        UOP_NU,
        FNC_INICIAL,
        FNC_FINAL
        )
        VALUES(
          $1, 
          $2, 
          $3
          )
          ON CONFLICT (UOP_NU, FNC_INICIAL) DO UPDATE SET
            FNC_FINAL = EXCLUDED.FNC_FINAL;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array} primaries
   */
  async DELETE_FROM_LOG_FAIXA_UOP(primaries) {
    const query = {
      text: `DELETE FROM correios.LOG_FAIXA_UOP
             WHERE UOP_NU = $1 AND FNC_INICIAL = $2;`,
      values: primaries,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  async CREATE_TABLE_ECT_PAIS() {
    const query = {
      text: `CREATE TABLE IF NOT EXISTS correios.ECT_PAIS(
        PAI_SG char(2) NOT NULL,
        PAI_SG_ALTERNATIVA char(3) NOT NULL,
        PAI_NO_PORTUGUES varchar(100) NOT NULL,
        PAI_NO_INGLES varchar(100) NOT NULL,
        PAI_NO_FRANCES varchar(100) NOT NULL,
        PAI_ABREVIATURA varchar(100) NOT NULL,
        PRIMARY KEY (PAI_SG)
      );
      COMMENT on column correios.ECT_PAIS.PAI_SG is 'Sigla do País';
      COMMENT on column correios.ECT_PAIS.PAI_SG_ALTERNATIVA is 'Sigla alternativa';
      `,
    };

    await this.DATABASE_CLIENT.query(query);
  }

  /**
   * @param {Array<string>} values
   */
  async INSERT_INTO_ECT_PAIS(values) {
    const query = {
      text: `INSERT INTO correios.ECT_PAIS(
        PAI_SG,
        PAI_SG_ALTERNATIVA,
        PAI_NO_PORTUGUES,
        PAI_NO_INGLES,
        PAI_NO_FRANCES,
        PAI_ABREVIATURA
        )
        VALUES(
          $1, 
          $2, 
          $3,
          $4,
          $5,
          $6
          )
          ON CONFLICT (PAI_SG) DO NOTHING;
          `,
      values,
    };

    await this.DATABASE_CLIENT.query(query);
  }
}
