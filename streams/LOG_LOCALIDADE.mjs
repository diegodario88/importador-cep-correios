import { createReadStream, existsSync, statSync } from "node:fs";
import { createInterface } from "node:readline";
import { AbstractStream } from "./abstract-stream.mjs";

export class LOG_LOCALIDADE_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../base-folder-files.mjs").BaseFolderOptions} options
   * @param {string} basePath
   * @returns {Promise<void>}
   */
  static async run(options, basePath) {
    const fileName = "LOG_LOCALIDADE.TXT";
    const filePath = `${basePath}/${fileName}`;

    if (!existsSync(filePath)) {
      return;
    }

    const fileSize = statSync(filePath).size;
    const fileLines = await this.getFileLines(filePath);
    const bar = options.multiBar.create(fileLines, 0, {
      filename: fileName,
    });

    await options.infra.CREATE_TABLE_LOG_LOCALIDADE();
    const readStream = createReadStream(filePath, "latin1");

    const readLine = createInterface({
      input: readStream,
      crlfDelay: Infinity,
    });

    for await (const line of readLine) {
      const data = line.split("@");
      const falsyToNull = (item) => (item ? item : null);
      await options.infra.INSERT_INTO_LOG_LOCALIDADE(data.map(falsyToNull));
      bar.increment();
    }

    options.fileSizeCount.push(fileSize);
    options.lineCount.push(fileLines);
  }
}
