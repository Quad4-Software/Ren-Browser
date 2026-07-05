// SPDX-License-Identifier: MIT
#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const packageDir = path.join(__dirname, "..", "..", "frontend", "node_modules", "micron-parser");
const packageJsonPath = path.join(packageDir, "package.json");

if (!fs.existsSync(packageDir)) {
  process.exit(0);
}

if (fs.existsSync(packageJsonPath)) {
  const existing = JSON.parse(fs.readFileSync(packageJsonPath, "utf8"));
  if (existing.main) {
    process.exit(0);
  }
}

const packageJson = {
  name: "micron-parser",
  version: "0.0.0",
  type: "module",
  main: "js/micron-parser.js",
  module: "js/micron-parser.js",
};

fs.writeFileSync(packageJsonPath, `${JSON.stringify(packageJson, null, 2)}\n`, "utf8");
