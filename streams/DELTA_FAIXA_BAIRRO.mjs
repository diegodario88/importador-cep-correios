import { createReadStream, statSync } from "node:fs";
import { cwd } from "node:process";
import { createInterface } from "node:readline";
import { AbstractStream } from "./abstract-stream.mjs";

export class DELTA_LOG_FAIXA_BAIRRO_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../delta-folder-files.mjs").DeltaFolderOptions} options
   * @returns {Promise<void>}
   */
  static async run(options) {
    const filePath = `${cwd()}/eDNE_Basico/eDNE_Delta_Basico_23011/Delimitado/DELTA_LOG_FAIXA_BAI.TXT`;
    const fileLines = await this.getFileLines(filePath);
    const fileSize = statSync(filePath).size;
    const bar = options.multiBar.create(fileLines, 0, {
      filename: filePath.split("/").pop(),
    });
    const readStream = createReadStream(filePath, "latin1");
    const readLine = createInterface({
      input: readStream,
      crlfDelay: Infinity,
    });

    streamLoop: for await (const line of readLine) {
      const data = line.split("@");
      const operation = data.pop();

      switch (operation) {
        case "INS":
          await options.infra.INSERT_INTO_LOG_FAIXA_BAIRRO(data);
          break;
        case "UPD":
          await options.infra.INSERT_INTO_LOG_FAIXA_BAIRRO(data);
          break;
        case "DEL":
          await options.infra.DELETE_FROM_LOG_FAIXA_BAIRRO([data[0], data[1]]);
          break;

        default:
          throw new Error("Operation not allowed");
      }

      bar.increment();
    }

    options.fileSizeCount.push(fileSize);
    options.lineCount.push(fileLines);
  }
}
