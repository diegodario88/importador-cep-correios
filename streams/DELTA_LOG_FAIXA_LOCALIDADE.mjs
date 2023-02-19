import { createReadStream, existsSync, statSync } from "node:fs";
import { createInterface } from "node:readline";
import { AbstractStream } from "./abstract-stream.mjs";

export class DELTA_LOG_FAIXA_LOCALIDADE_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../delta-folder-files.mjs").DeltaFolderOptions} options
   * @param {string} basePath
   * @returns {Promise<void>}
   */
  static async run(options, basePath) {
    const fileName = "DELTA_LOG_FAIXA_LOCALIDADE.TXT";
    const filePath = `${basePath}/${fileName}`;

    if (!existsSync(filePath)) {
      return;
    }

    const fileSize = statSync(filePath).size;
    const fileLines = await this.getFileLines(filePath);
    const bar = options.multiBar.create(fileLines, 0, {
      filename: fileName,
    });

    const readStream = createReadStream(filePath, "latin1");
    const readLine = createInterface({
      input: readStream,
      crlfDelay: Infinity,
    });

    streamLoop: for await (const line of readLine) {
      const data = line.split("@");
      const tipoFaixa = data.pop();
      const operation = data.pop();
      data.push(tipoFaixa);

      switch (operation) {
        case "INS":
          await options.infra.INSERT_INTO_LOG_FAIXA_LOCALIDADE(data);
          break;
        case "UPD":
          await options.infra.INSERT_INTO_LOG_FAIXA_LOCALIDADE(data);
          break;
        case "DEL":
          await options.infra.DELETE_FROM_LOG_FAIXA_LOCALIDADE([
            data[0],
            data[1],
            tipoFaixa,
          ]);
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
