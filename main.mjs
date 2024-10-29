import * as dotenv from "dotenv";
import { InfrastructureService } from "./infra.mjs";
import { BaseFolderFiles } from "./base-folder-files.mjs";
import { MultiBar, Presets } from "cli-progress";
import { DeltaFolderFiles } from "./delta-folder-files.mjs";

console.time("import-execution-time");
dotenv.config();
const adder = (numbers) => numbers.reduce((acc, current) => acc + current, 0);
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const infra = new InfrastructureService();
const fileSizeCount = [0];
const lineCount = [0];
const baseMultiBar = new MultiBar(
  {
    clearOnComplete: false,
    hideCursor: true,
    fps: 60,
    format: " {bar} {percentage}% of {total} | {filename} ",
    barsize: 30,
    autopadding: true,
    forceRedraw: true,
    stopOnComplete: true,
    formatValue: (v) => v.toLocaleString("pt-BR"),
  },
  Presets.rect,
);
const deltaMultiBar = new MultiBar(
  {
    clearOnComplete: false,
    hideCursor: true,
    fps: 60,
    format: " {bar} {percentage}% of {total} | {filename} ",
    barsize: 30,
    autopadding: true,
    forceRedraw: true,
    stopOnComplete: true,
    formatValue: (v) => v.toLocaleString("pt-BR"),
  },
  Presets.rect,
);

try {
  await infra.connectToDatabase();
  await infra.createCorreiosSchema();

  const baseFolder = new BaseFolderFiles({
    fileSizeCount,
    infra,
    lineCount,
    multiBar: baseMultiBar,
  });
  console.log("Importing files ...");
  await baseFolder.process();
  await sleep(700);

  const deltaFolder = new DeltaFolderFiles({
    fileSizeCount,
    infra,
    lineCount,
    multiBar: deltaMultiBar,
  });
  console.log("Updating delta ...");
  await deltaFolder.process();
  await sleep(700);

  /**
   * CONCLUSION REPORTS
   */
  const totalInserted = await infra.getTotalRecords();
  const totalCEPS = await infra.getTotalCEPS();
  const totalFileSize = adder(fileSizeCount);
  const fileSizeInMBytes = totalFileSize / 1024 / 1024;
  const totalLines = adder(lineCount);

  console.log(`\n \nTotal lines read: ${totalLines.toLocaleString("pt-BR")}`);
  console.log(`Total files throughput: ${fileSizeInMBytes.toFixed(2)} MB`);
  console.log(`Total of records: ${totalInserted.toLocaleString("pt-BR")}`);
  console.log(`Total of CEPS: ${totalCEPS.toLocaleString("pt-BR")}`);
} catch (error) {
  console.error(error);
} finally {
  await infra.disconnectToDatabase();
}
console.timeEnd("import-execution-time");
