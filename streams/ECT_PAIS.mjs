import { createReadStream, existsSync, statSync } from "node:fs";
import { createInterface } from "node:readline";
import { AbstractStream } from "./abstract-stream.mjs";

export class ECT_PAIS_STREAM extends AbstractStream {
  /**
   *
   * @param {import("../base-folder-files.mjs").BaseFolderOptions} options
   * @param {string} basePath
   * @returns {Promise<void>}
   */
  static async run(options, basePath) {
    const fileName = "ECT_PAIS.TXT";
    const filePath = `${basePath}/${fileName}`;

    if (!existsSync(filePath)) {
      return;
    }

    const fileSize = statSync(filePath).size;
    const fileLines = await this.getFileLines(filePath);
    const bar = options.multiBar.create(fileLines, 0, {
      filename: fileName,
    });

    await options.infra.CREATE_TABLE_ECT_PAIS();
    const readStream = createReadStream(filePath, "latin1");

    const readLine = createInterface({
      input: readStream,
      crlfDelay: Infinity,
    });

    for await (const line of readLine) {
      const data = line.split("@");
      await options.infra.INSERT_INTO_ECT_PAIS(data);
      bar.increment();
    }

    options.fileSizeCount.push(fileSize);
    options.lineCount.push(fileLines);
  }
}
