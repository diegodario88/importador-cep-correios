import { MultiBar } from "cli-progress";
import { cwd } from "node:process";
import { InfrastructureService } from "./infra.mjs";
import { ECT_PAIS_STREAM } from "./streams/ECT_PAIS.mjs";
import { LOG_BAIRRO_STREAM } from "./streams/LOG_BAIRRO.mjs";
import { LOG_CPC_STREAM } from "./streams/LOG_CPC.mjs";
import { LOG_FAIXA_BAIRRO_STREAM } from "./streams/LOG_FAIXA_BAIRRO.mjs";
import { LOG_FAIXA_CPC_STREAM } from "./streams/LOG_FAIXA_CPC.mjs";
import { LOG_FAIXA_LOCALIDADE_STREAM } from "./streams/LOG_FAIXA_LOCALIDADE.mjs";
import { LOG_FAIXA_UF_STREAM } from "./streams/LOG_FAIXA_UF.mjs";
import { LOG_FAIXA_UOP_STREAM } from "./streams/LOG_FAIXA_UOP.mjs";
import { LOG_GRANDE_USUARIO_STREAM } from "./streams/LOG_GRANDE_USUARIO.mjs";
import { LOG_LOCALIDADE_STREAM } from "./streams/LOG_LOCALIDADE.mjs";
import { LOG_LOGRADOURO_STREAM } from "./streams/LOG_LOGRADOURO.mjs";
import { LOG_NUM_SEC_STREAM } from "./streams/LOG_NUM_SEC.mjs";
import { LOG_UNID_OPER_STREAM } from "./streams/LOG_UNID_OPER.mjs";
import { LOG_VAR_BAI_STREAM } from "./streams/LOG_VAR_BAI.mjs";
import { LOG_VAR_LOC_STREAM } from "./streams/LOG_VAR_LOC.mjs";
import { LOG_VAR_LOG_STREAM } from "./streams/LOG_VAR_LOG.mjs";

/**
 * @typedef BaseFolderOptions
 * @type {object}
 * @property {InfrastructureService} infra - InfrastructureService.
 * @property {Array<number>} lineCount - Counter of lines reading.
 * @property {Array<number>} fileSizeCount - Counter of fileSize.
 * @property {MultiBar} multiBar - MultiBar.
 */

export class BaseFolderFiles {
  _options;

  /**
   * @param {BaseFolderOptions} options
   */
  constructor(options) {
    this._options = options;
  }

  async process() {
    /**
     * Caminho onde se encontram os arquivos `.TXT` do modelo básico
     * delimitado por `@`
     */
    const basePath = `${cwd()}/eDNE/basico`;

    await Promise.all([
      LOG_FAIXA_UF_STREAM.run(this._options, basePath),
      LOG_LOCALIDADE_STREAM.run(this._options, basePath),
      LOG_VAR_LOC_STREAM.run(this._options, basePath),
      LOG_FAIXA_LOCALIDADE_STREAM.run(this._options, basePath),
      LOG_BAIRRO_STREAM.run(this._options, basePath),
      LOG_VAR_BAI_STREAM.run(this._options, basePath),
      LOG_FAIXA_BAIRRO_STREAM.run(this._options, basePath),
      LOG_CPC_STREAM.run(this._options, basePath),
      LOG_FAIXA_CPC_STREAM.run(this._options, basePath),
      LOG_LOGRADOURO_STREAM.run(this._options, basePath),
      LOG_VAR_LOG_STREAM.run(this._options, basePath),
      LOG_NUM_SEC_STREAM.run(this._options, basePath),
      LOG_GRANDE_USUARIO_STREAM.run(this._options, basePath),
      LOG_UNID_OPER_STREAM.run(this._options, basePath),
      LOG_FAIXA_UOP_STREAM.run(this._options, basePath),
      ECT_PAIS_STREAM.run(this._options, basePath),
    ]);
  }
}
