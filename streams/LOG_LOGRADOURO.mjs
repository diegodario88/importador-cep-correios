import { createReadStream, existsSync, statSync } from "node:fs";
import { createInterface } from "node:readline";
import { federative_units } from "../federative-units.mjs";
import { AbstractStream } from "./abstract-stream.mjs";

export class LOG_LOGRADOURO_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../base-folder-files.mjs").BaseFolderOptions} options
   * @param {string} basePath
   * @returns {Promise<void>}
   */
  static async run(options, basePath) {
    await options.infra.CREATE_TABLE_LOG_LOGRADOURO();

    for (const { sigla } of federative_units) {
      const fileName = `LOG_LOGRADOURO_${sigla}.TXT`;
      const filePath = `${basePath}/${fileName}`;

      if (!existsSync(filePath)) {
        continue;
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
