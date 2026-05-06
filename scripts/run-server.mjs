import { spawn } from "node:child_process";

const options = parseArgs(process.argv.slice(2));

if (options.help) {
  printHelp();
  process.exit(0);
}

const env = {
  ...process.env,
};

if (options.preset) {
  env.GOCMS_PRESET = options.preset;
}
if (options.runtime) {
  env.GOCMS_RUNTIME_PROFILE = options.runtime;
}
if (options.storage) {
  env.GOCMS_STORAGE_PROFILE = options.storage;
}
if (options.bind) {
  env.APP_BIND = options.bind;
}
if (options.plugins) {
  env.GOCMS_PLUGIN_SET = options.plugins;
}
if (options.sitePackageDir) {
  env.GOCMS_SITE_PACKAGE_DIR = options.sitePackageDir;
}

console.log("Starting GoCMS local server with:");
console.log(`- preset: ${env.GOCMS_PRESET ?? "(default)"}`);
console.log(`- runtime: ${env.GOCMS_RUNTIME_PROFILE ?? "(from preset/default)"}`);
console.log(`- storage: ${env.GOCMS_STORAGE_PROFILE ?? "(from preset/default)"}`);
console.log(`- bind: ${env.APP_BIND ?? "(from preset/default)"}`);
if (env.GOCMS_PLUGIN_SET) {
  console.log(`- plugins: ${env.GOCMS_PLUGIN_SET}`);
}
if (env.GOCMS_SITE_PACKAGE_DIR) {
  console.log(`- site package: ${env.GOCMS_SITE_PACKAGE_DIR}`);
}

const child = spawn("go", ["run", "./cmd/server"], {
  cwd: process.cwd(),
  stdio: "inherit",
  env,
});

child.on("exit", (code) => {
  process.exit(code ?? 0);
});

function parseArgs(args) {
  const result = {
    preset: "",
    runtime: "",
    storage: "",
    bind: "",
    plugins: "",
    sitePackageDir: "",
    help: false,
  };

  for (let i = 0; i < args.length; i += 1) {
    const arg = args[i];
    switch (arg) {
      case "--preset":
        result.preset = args[++i] ?? "";
        break;
      case "--runtime":
        result.runtime = args[++i] ?? "";
        break;
      case "--storage":
        result.storage = args[++i] ?? "";
        break;
      case "--bind":
        result.bind = args[++i] ?? "";
        break;
      case "--plugins":
        result.plugins = args[++i] ?? "";
        break;
      case "--site-package-dir":
        result.sitePackageDir = args[++i] ?? "";
        break;
      case "--help":
      case "-h":
        result.help = true;
        break;
      default:
        console.error(`Unknown argument: ${arg}`);
        printHelp();
        process.exit(1);
    }
  }

  return result;
}

function printHelp() {
  console.log(`Usage: node ./scripts/run-server.mjs [options]

Options:
  --preset <id>             Set GOCMS_PRESET
  --runtime <profile>       Set GOCMS_RUNTIME_PROFILE
  --storage <profile>       Set GOCMS_STORAGE_PROFILE
  --bind <host:port>        Set APP_BIND
  --plugins <csv>           Set GOCMS_PLUGIN_SET
  --site-package-dir <dir>  Set GOCMS_SITE_PACKAGE_DIR
  --help                    Show this help
`);
}
