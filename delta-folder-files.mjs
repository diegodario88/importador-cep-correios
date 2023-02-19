import { MultiBar } from "cli-progress";
import { InfrastructureService } from "./infra.mjs";
import { DELTA_LOG_FAIXA_BAIRRO_STREAM } from "./streams/DELTA_FAIXA_BAIRRO.mjs";
import { DELTA_LOG_BAIRRO_STREAM } from "./streams/DELTA_LOG_BAIRRO.mjs";
import { DELTA_LOG_CPC_STREAM } from "./streams/DELTA_LOG_CPC.mjs";
import { DELTA_LOG_FAIXA_CPC_STREAM } from "./streams/DELTA_LOG_FAIXA_CPC.mjs";
import { DELTA_LOG_FAIXA_LOCALIDADE_STREAM } from "./streams/DELTA_LOG_FAIXA_LOCALIDADE.mjs";
import { DELTA_LOG_FAIXA_UOP_STREAM } from "./streams/DELTA_LOG_FAIXA_UOP.mjs";
import { DELTA_LOG_GRANDE_USUARIO_STREAM } from "./streams/DELTA_LOG_GRANDE_USUARIO.mjs";
import { DELTA_LOG_LOCALIDADE_STREAM } from "./streams/DELTA_LOG_LOCALIDADE.mjs";
import { DELTA_LOG_LOGRADOURO_STREAM } from "./streams/DELTA_LOG_LOGRADOURO.mjs";
import { DELTA_LOG_NUM_SEC_STREAM } from "./streams/DELTA_LOG_NUM_SEC.mjs";
import { DELTA_LOG_UNID_OPER_STREAM } from "./streams/DELTA_LOG_UNID_OPER.mjs";

/**
 * @typedef DeltaFolderOptions
 * @type {object}
 * @property {InfrastructureService} infra - InfrastructureService.
 * @property {Array<number>} lineCount - Counter of lines reading.
 * @property {Array<number>} fileSizeCount - Counter of fileSize.
 * @property {MultiBar} multiBar - MultiBar.
 */

export class DeltaFolderFiles {
  _options;

  /**
   * @param {DeltaFolderOptions} options
   */
  constructor(options) {
    this._options = options;
  }

  async process() {
    await Promise.all([
      DELTA_LOG_BAIRRO_STREAM.run(this._options),
      DELTA_LOG_CPC_STREAM.run(this._options),
      DELTA_LOG_FAIXA_BAIRRO_STREAM.run(this._options),
      DELTA_LOG_FAIXA_CPC_STREAM.run(this._options),
      DELTA_LOG_FAIXA_LOCALIDADE_STREAM.run(this._options),
      DELTA_LOG_FAIXA_UOP_STREAM.run(this._options),
      DELTA_LOG_GRANDE_USUARIO_STREAM.run(this._options),
      DELTA_LOG_LOCALIDADE_STREAM.run(this._options),
      DELTA_LOG_LOGRADOURO_STREAM.run(this._options),
      DELTA_LOG_NUM_SEC_STREAM.run(this._options),
      DELTA_LOG_UNID_OPER_STREAM.run(this._options),
    ]);
  }
}
