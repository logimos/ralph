import { getPackageVersion } from "../version.js";

export type HealthResult = {
  ok: true;
  version: string;
  service: "layers";
};

export function runCli(argv: string[]): { code: number; stdout: string; stderr: string } {
  if (argv[0] === "v1" && argv[1] === "health") {
    const body: HealthResult = {
      ok: true,
      version: getPackageVersion(),
      service: "layers",
    };
    return { code: 0, stdout: `${JSON.stringify(body)}\n`, stderr: "" };
  }

  const stderr =
    "Usage: layers v1 health\nSee docs/LAYERS_SPEC.md — more commands in later phases.\n";
  return { code: 1, stdout: "", stderr };
}
