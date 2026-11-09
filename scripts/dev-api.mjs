import { spawn } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const apiDirectory = path.join(repositoryRoot, "apps", "api");
const isWindows = process.platform === "win32";
const airCommand = isWindows ? "air.exe" : "air";
const binary = isWindows ? "./tmp/main.exe" : "./tmp/main";
const buildCommand = isWindows
  ? "go build -o ./tmp/main.exe ./cmd/server"
  : "go build -o ./tmp/main ./cmd/server";

const air = spawn(
  airCommand,
  ["-build.cmd", buildCommand, "-build.entrypoint", binary],
  {
    cwd: apiDirectory,
    stdio: "inherit",
    shell: false,
  },
);

air.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 1);
  }
});
