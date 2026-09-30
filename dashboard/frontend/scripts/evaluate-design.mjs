#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDir, '..');
const repoRoot = path.resolve(frontendRoot, '../..');
const contractPath = path.resolve(frontendRoot, '../design-system/design-contract.json');
const contract = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
const jsonOutput = process.argv.includes('--json');

function readRepoFile(relativePath) {
  return fs.readFileSync(path.resolve(repoRoot, relativePath), 'utf8');
}

function walk(directory) {
  if (!fs.existsSync(directory)) return [];
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) return walk(fullPath);
    return entry.isFile() && /\.(tsx|ts)$/.test(entry.name) ? [fullPath] : [];
  });
}

const results = [];
let score = 100;
let blockers = 0;

function pass(id, message) {
  results.push({ id, status: 'pass', message });
}

function warn(id, message, penalty = 0) {
  score -= penalty;
  results.push({ id, status: 'warn', message });
}

function fail(id, message, penalty = 10) {
  score -= penalty;
  blockers += 1;
  results.push({ id, status: 'fail', message });
}

for (const relativePath of contract.requiredFiles) {
  if (fs.existsSync(path.resolve(repoRoot, relativePath))) {
    pass(`file:${relativePath}`, 'required file exists');
  } else {
    fail(`file:${relativePath}`, 'required file is missing');
  }
}

const css = readRepoFile('dashboard/frontend/app/globals.css');
for (const token of contract.requiredCssTokens) {
  if (css.includes(`${token}:`)) {
    pass(`token:${token}`, `semantic token ${token} is declared`);
  } else {
    fail(`token:${token}`, `semantic token ${token} is missing`);
  }
}

for (const selector of contract.requiredCssSelectors) {
  if (css.includes(selector)) {
    pass(`selector:${selector}`, `component recipe ${selector} exists`);
  } else {
    fail(`selector:${selector}`, `component recipe ${selector} is missing`);
  }
}

for (const check of contract.forbiddenCssPatterns) {
  const expression = new RegExp(check.pattern);
  if (expression.test(css)) {
    fail(`forbidden:${check.pattern}`, check.reason);
  } else {
    pass(`forbidden:${check.pattern}`, 'forbidden canvas treatment is absent');
  }
}

for (const check of contract.discouragedCssPatterns) {
  const expression = new RegExp(check.pattern);
  if (expression.test(css)) {
    warn(`discouraged:${check.pattern}`, `${check.reason} Keep this limited to loading states.`, 2);
  } else {
    pass(`discouraged:${check.pattern}`, 'no discouraged gradient usage found');
  }
}

for (const check of contract.requiredMarkers) {
  const content = readRepoFile(check.file);
  if (content.includes(check.pattern)) {
    pass(`marker:${check.file}:${check.pattern}`, check.reason);
  } else {
    fail(`marker:${check.file}:${check.pattern}`, check.reason);
  }
}

const componentFiles = walk(path.join(frontendRoot, 'app')).concat(
  walk(path.join(frontendRoot, 'components')),
);
for (const check of contract.forbiddenComponentPatterns) {
  const expression = new RegExp(check.pattern);
  const matches = componentFiles.filter((file) => expression.test(fs.readFileSync(file, 'utf8')));
  if (matches.length > 0) {
    fail(
      `component:${check.pattern}`,
      `${check.reason} Found in ${matches.map((file) => path.relative(repoRoot, file)).join(', ')}.`,
      4,
    );
  } else {
    pass(`component:${check.pattern}`, 'component colors use semantic classes');
  }
}

score = Math.max(0, Math.min(100, score));
const status = blockers === 0 && score >= 85 ? 'PASS' : 'FAIL';
const report = {
  contract: contract.name,
  version: contract.version,
  status,
  score,
  blockers,
  checks: results,
};

if (jsonOutput) {
  console.log(JSON.stringify(report, null, 2));
} else {
  console.log(`Sloper design contract: ${status}`);
  console.log(`Score: ${score}/100 · blockers: ${blockers}`);
  for (const result of results) {
    const marker = result.status === 'pass' ? '✓' : result.status === 'warn' ? '!' : '✗';
    console.log(`${marker} ${result.id}: ${result.message}`);
  }
  console.log('\nStatic checks cannot judge composition. Run the Chrome MCP visual pass at desktop and mobile sizes.');
}

process.exitCode = status === 'PASS' ? 0 : 1;
