import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

export class AbstractStream {
  /**
   * @param {string} filePath
   */
  static async getFileLines(filePath) {
    const readStream = createReadStream(filePath, "latin1");
    const readLine = createInterface({
      input: readStream,
      crlfDelay: Infinity,
    });

    let lineCounter = 0;
    for await (const line of readLine) {
      lineCounter++;
    }

    return lineCounter;
  }
}
