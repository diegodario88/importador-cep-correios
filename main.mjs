import * as dotenv from "dotenv";
import { InfrastructureService } from "./infra.mjs";
import { BaseFolderFiles } from "./base-folder-files.mjs";
import { MultiBar, Presets } from "cli-progress";
import { DeltaFolderFiles } from "./delta-folder-files.mjs";

console.time("import-execution-time");
dotenv.config();
const adder = (numbers) => numbers.reduce((acc, current) => acc + current, 0);
const infra = new InfrastructureService();
const fileSizeCount = [0];
const lineCount = [0];
const multiBar = new MultiBar(
  {
    clearOnComplete: false,
    hideCursor: true,
    fps: 60,
    format: " {bar} {percentage}% of {total} | {filename} ",
    autopadding: true,
    forceRedraw: true,
    barsize: 30,
    formatValue: (v) => v.toLocaleString("pt-BR"),
  },
  Presets.rect
);

try {
  await infra.connectToDatabase();
  await infra.createCorreiosSchema();

  const baseFolder = new BaseFolderFiles({
    fileSizeCount,
    infra,
    lineCount,
    multiBar,
  });
  console.log("Importing files ...");
  await baseFolder.process();

  const deltaFolder = new DeltaFolderFiles({
    fileSizeCount,
    infra,
    lineCount,
    multiBar,
  });
  console.log("Updating delta ...");
  await deltaFolder.process();

  /**
   * CONCLUSION REPORTS
   */
  const totalInserted = await infra.getTotalRecords();
  const totalCEPS = await infra.getTotalCEPS();
  const totalFileSize = adder(fileSizeCount);
  const fileSizeInMBytes = totalFileSize / 1024 / 1024;
  const totalLines = adder(lineCount);

  console.log(
    `\n \nTotal lines read: ${totalLines.toLocaleString("pt-BR", {
      minimumFractionDigits: 3,
    })}`
  );
  console.log(`Total files throughput: ${fileSizeInMBytes.toFixed(2)} MB`);
  console.log(
    `Total of records: ${totalInserted.toLocaleString("pt-BR", {
      minimumFractionDigits: 3,
    })}`
  );
  console.log(
    `Total of CEPS: ${totalCEPS.toLocaleString("pt-BR", {
      minimumFractionDigits: 3,
    })}`
  );
} catch (error) {
  console.error(error);
} finally {
  await infra.disconnectToDatabase();
}
console.timeEnd("import-execution-time");
