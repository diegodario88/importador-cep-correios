import { createReadStream, statSync } from "node:fs";
import { cwd } from "node:process";
import { createInterface } from "node:readline";
import { federative_units } from "../federative-units.mjs";
import { AbstractStream } from "./abstract-stream.mjs";

export class LOG_LOGRADOURO_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../base-folder-files.mjs").BaseFolderOptions} options
   * @returns {Promise<void>}
   */
  static async run(options) {
    await options.infra.CREATE_TABLE_LOG_LOGRADOURO();

    for (const uf of federative_units) {
      const filePath = `${cwd()}/eDNE_Basico/eDNE_Basico_23012/Delimitado/LOG_LOGRADOURO_${
        uf.sigla
      }.TXT`;
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

      for await (const line of readLine) {
        const data = line.split("@");
        const falsyToNull = (item) => (item ? item : null);
        await options.infra.INSERT_INTO_LOG_LOGRADOURO(data.map(falsyToNull));
        bar.increment();
      }

      options.fileSizeCount.push(fileSize);
      options.lineCount.push(fileLines);
    }
  }
}
