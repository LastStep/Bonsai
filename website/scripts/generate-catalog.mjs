#!/usr/bin/env node

/**
 * generate-catalog.mjs
 *
 * Reads catalog/meta.yaml and catalog/agents/agent.yaml source files,
 * then updates the docs site catalog pages (between marker comments)
 * and writes website/public/catalog.json.
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import yaml from 'js-yaml';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const WEBSITE_DIR = path.resolve(__dirname, '..');
const CATALOG_DIR = path.resolve(WEBSITE_DIR, '..', 'catalog');
const DOCS_DIR = path.join(WEBSITE_DIR, 'src', 'content', 'docs', 'catalog');
const PUBLIC_DIR = path.join(WEBSITE_DIR, 'public');

// Workflows that have slash commands (from internal/generate/generate.go:CuratedSlashWorkflows)
const CURATED_SLASH_WORKFLOWS = new Set([
  'planning',
  'code-review',
  'pr-review',
  'security-audit',
  'issue-to-implementation',
  'test-plan',
  'plan-execution',
]);

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function readYamlDir(dir, filename = 'meta.yaml') {
  const items = [];
  if (!fs.existsSync(dir)) return items;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const yamlPath = path.join(dir, entry.name, filename);
    if (!fs.existsSync(yamlPath)) continue;
    const data = yaml.load(fs.readFileSync(yamlPath, 'utf8'));
    items.push(data);
  }
  return items.sort((a, b) => a.name.localeCompare(b.name));
}

function displayName(name) {
  return name
    .split('-')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function formatAgents(agents) {
  if (agents === 'all') return 'all';
  if (Array.isArray(agents)) return agents.sort().join(', ');
  return String(agents);
}

function formatRequired(item) {
  if (item.name === 'routine-check') return 'Auto-managed';
  if (item.required === 'all') return 'Yes (all)';
  if (item.required) return `Yes (${formatAgents(item.required)})`;
  return 'No';
}

/**
 * Replace content between marker comments in a file.
 * Returns true if the file was modified.
 */
function replaceMarkerContent(filePath, startMarker, endMarker, newContent) {
  const text = fs.readFileSync(filePath, 'utf8');
  const startIdx = text.indexOf(startMarker);
  const endIdx = text.indexOf(endMarker);
  if (startIdx === -1 || endIdx === -1) {
    console.warn(`  WARNING: Markers not found in ${path.basename(filePath)}: ${startMarker}`);
    return false;
  }
  const before = text.slice(0, startIdx + startMarker.length);
  const after = text.slice(endIdx);
  const updated = before + '\n' + newContent + '\n' + after;
  if (updated !== text) {
    fs.writeFileSync(filePath, updated, 'utf8');
    return true;
  }
  return false;
}

/**
 * Update the frontmatter description line to reflect accurate count.
 * Looks for a pattern like "All N Bonsai {type}" and replaces N.
 */
function updateFrontmatterCount(filePath, count, typeWord) {
  const text = fs.readFileSync(filePath, 'utf8');
  const pattern = new RegExp(`(All )\\d+( Bonsai ${typeWord})`, 'i');
  if (pattern.test(text)) {
    const updated = text.replace(pattern, `$1${count}$2`);
    if (updated !== text) {
      fs.writeFileSync(filePath, updated, 'utf8');
    }
  }
}

// ---------------------------------------------------------------------------
// Load all catalog data
// ---------------------------------------------------------------------------

const skills = readYamlDir(path.join(CATALOG_DIR, 'skills'));
const workflows = readYamlDir(path.join(CATALOG_DIR, 'workflows'));
const protocols = readYamlDir(path.join(CATALOG_DIR, 'protocols'));
const sensors = readYamlDir(path.join(CATALOG_DIR, 'sensors'));
const routines = readYamlDir(path.join(CATALOG_DIR, 'routines'));
const agents = readYamlDir(path.join(CATALOG_DIR, 'agents'), 'agent.yaml');

console.log('Catalog loaded:');
console.log(`  Agents:    ${agents.length}`);
console.log(`  Skills:    ${skills.length}`);
console.log(`  Workflows: ${workflows.length}`);
console.log(`  Protocols: ${protocols.length}`);
console.log(`  Sensors:   ${sensors.length}`);
console.log(`  Routines:  ${routines.length}`);
console.log(`  Total:     ${agents.length + skills.length + workflows.length + protocols.length + sensors.length + routines.length}`);
console.log();

// ---------------------------------------------------------------------------
// Generate: Skills table
// ---------------------------------------------------------------------------

{
  const lines = [
    '| Name | Description | Compatible Agents |',
    '|:-----|:-----------|:-----------------|',
  ];
  for (const s of skills) {
    lines.push(`| ${s.name} | ${s.description} | ${formatAgents(s.agents)} |`);
  }
  const filePath = path.join(DOCS_DIR, 'skills.mdx');
  const changed = replaceMarkerContent(filePath, '{/* CATALOG-TABLE-START */}', '{/* CATALOG-TABLE-END */}', lines.join('\n'));
  updateFrontmatterCount(filePath, skills.length, 'skills');
  console.log(`skills.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Workflows table
// ---------------------------------------------------------------------------

{
  const lines = [
    '| Name | Description | Compatible Agents | Slash Command |',
    '|:-----|:-----------|:-----------------|:-------------|',
  ];
  for (const w of workflows) {
    const slash = CURATED_SLASH_WORKFLOWS.has(w.name) ? `\`/${w.name}\`` : '';
    lines.push(`| ${w.name} | ${w.description} | ${formatAgents(w.agents)} | ${slash} |`);
  }
  const filePath = path.join(DOCS_DIR, 'workflows.mdx');
  const changed = replaceMarkerContent(filePath, '{/* CATALOG-TABLE-START */}', '{/* CATALOG-TABLE-END */}', lines.join('\n'));
  updateFrontmatterCount(filePath, workflows.length, 'workflows');
  console.log(`workflows.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Protocols table
// ---------------------------------------------------------------------------

{
  const lines = [
    '| Name | Description | Required |',
    '|:-----|:-----------|:---------|',
  ];
  for (const p of protocols) {
    lines.push(`| ${p.name} | ${p.description} | ${formatRequired(p)} |`);
  }
  const filePath = path.join(DOCS_DIR, 'protocols.mdx');
  const changed = replaceMarkerContent(filePath, '{/* CATALOG-TABLE-START */}', '{/* CATALOG-TABLE-END */}', lines.join('\n'));
  updateFrontmatterCount(filePath, protocols.length, 'protocols');
  console.log(`protocols.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Sensors table
// ---------------------------------------------------------------------------

{
  const lines = [
    '| Name | Description | Event | Matcher | Compatible Agents | Required |',
    '|:-----|:-----------|:------|:--------|:-----------------|:---------|',
  ];
  for (const s of sensors) {
    lines.push(
      `| ${s.name} | ${s.description} | ${s.event} | ${s.matcher || ''} | ${formatAgents(s.agents)} | ${formatRequired(s)} |`
    );
  }
  const filePath = path.join(DOCS_DIR, 'sensors.mdx');
  const changed = replaceMarkerContent(filePath, '{/* CATALOG-TABLE-START */}', '{/* CATALOG-TABLE-END */}', lines.join('\n'));
  updateFrontmatterCount(filePath, sensors.length, 'sensors');
  console.log(`sensors.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Routines table
// ---------------------------------------------------------------------------

{
  const lines = [
    '| Name | Description | Frequency | Compatible Agents |',
    '|:-----|:-----------|:----------|:-----------------|',
  ];
  for (const r of routines) {
    lines.push(`| ${r.name} | ${r.description} | ${r.frequency} | ${formatAgents(r.agents)} |`);
  }
  const filePath = path.join(DOCS_DIR, 'routines.mdx');
  const changed = replaceMarkerContent(filePath, '{/* CATALOG-TABLE-START */}', '{/* CATALOG-TABLE-END */}', lines.join('\n'));
  updateFrontmatterCount(filePath, routines.length, 'routines');
  console.log(`routines.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Agent Types page
// ---------------------------------------------------------------------------

{
  // Sort: tech-lead first, then alphabetical
  const techLead = agents.find((a) => a.name === 'tech-lead');
  const codeAgents = agents.filter((a) => a.name !== 'tech-lead').sort((a, b) => a.name.localeCompare(b.name));

  function agentDefaultsList(defaults, category) {
    const items = defaults?.[category];
    if (!items || items.length === 0) {
      // Special case: tech-lead routines
      if (category === 'routines' && techLead) {
        return `_(none by default — all ${routines.length} catalog routines are compatible)_`;
      }
      return '_(none)_';
    }
    return items.join(', ');
  }

  function generateAgentSection(agent, heading) {
    const dn = agent.display_name || displayName(agent.name);
    const lines = [];
    lines.push(`### ${dn}`);
    lines.push('');
    lines.push(`**Description:** ${agent.description}`);
    lines.push('');

    // Generate "Best for" from the description
    const bestFor = generateBestFor(agent);
    lines.push(`**Best for:** ${bestFor}`);
    lines.push('');

    if (agent.name === 'tech-lead') {
      lines.push('<Aside type="caution">');
      lines.push(
        'The Tech Lead never writes application code directly. It creates plans, dispatches them to code agents via worktree-isolated subagents, and reviews the results.'
      );
      lines.push('</Aside>');
      lines.push('');
    }

    lines.push('| Category | Defaults |');
    lines.push('|:---------|:---------|');

    const isThisAgent = agent.name === techLead?.name;

    for (const cat of ['skills', 'workflows', 'protocols', 'sensors', 'routines']) {
      let val;
      const items = agent.defaults?.[cat];
      if (!items || items.length === 0) {
        if (cat === 'routines' && isThisAgent) {
          val = `_(none by default — all ${routines.length} catalog routines are compatible)_`;
        } else {
          val = '_(none)_';
        }
      } else {
        val = items.join(', ');
      }
      lines.push(`| **${cat.charAt(0).toUpperCase() + cat.slice(1)}** | ${val} |`);
    }

    return lines.join('\n');
  }

  function generateBestFor(agent) {
    const bestForMap = {
      'tech-lead':
        'Project leads who need an agent that plans features, dispatches work to code agents, and reviews their output.',
      backend: 'API development, database work, server-side business logic, and backend infrastructure.',
      frontend: 'UI component development, state management, CSS/styling, and frontend architecture.',
      fullstack:
        'Features that span the entire stack, when you want one agent handling frontend through database.',
      devops: 'Terraform, Docker, Kubernetes, CI/CD pipelines, and deployment configuration.',
      security:
        'Security audits, vulnerability scanning, dependency review, and auth pattern enforcement.',
    };
    return bestForMap[agent.name] || agent.description;
  }

  const sections = [];

  // Orchestrator section
  sections.push('## The Orchestrator');
  sections.push('');
  sections.push(generateAgentSection(techLead, '###'));
  sections.push('');
  sections.push('---');
  sections.push('');
  sections.push('## Code Agents');
  sections.push('');
  sections.push(
    'Code agents execute plans created by the Tech Lead. Each has domain-specific skills and a focused scope.'
  );

  for (let i = 0; i < codeAgents.length; i++) {
    sections.push('');
    sections.push(generateAgentSection(codeAgents[i], '###'));
    if (i < codeAgents.length - 1) {
      sections.push('');
      sections.push('---');
    }
  }

  const filePath = path.join(DOCS_DIR, 'agent-types.mdx');
  const changed = replaceMarkerContent(
    filePath,
    '{/* AGENT-DEFAULTS-START */}',
    '{/* AGENT-DEFAULTS-END */}',
    sections.join('\n')
  );
  console.log(`agent-types.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: Overview page counts
// ---------------------------------------------------------------------------

{
  const filePath = path.join(DOCS_DIR, 'overview.mdx');
  const countLines = [
    '<CardGrid>',
    '  <Card title="Agent Types" icon="rocket">',
    `    **${agents.length} agents** — ${agents.map((a) => a.display_name || displayName(a.name)).join(', ')}. Each has a specialized identity, defaults, and role.`,
    '',
    '    [Browse agent types](/Bonsai/catalog/agent-types/)',
    '  </Card>',
    '  <Card title="Skills" icon="open-book">',
    `    **${skills.length} skills** — Domain knowledge and standards. Coding conventions, API design, testing strategy, infrastructure patterns, and more.`,
    '',
    '    [Browse skills](/Bonsai/catalog/skills/)',
    '  </Card>',
    '  <Card title="Workflows" icon="list-format">',
    `    **${workflows.length} workflows** — Step-by-step procedures for planning, code review, security audits, PR review, reporting, and implementation.`,
    '',
    '    [Browse workflows](/Bonsai/catalog/workflows/)',
    '  </Card>',
    '  <Card title="Protocols" icon="warning">',
    `    **${protocols.length} protocols** — Hard rules that every agent must follow. Memory management, scope boundaries, security enforcement, and session startup.`,
    '',
    '    [Browse protocols](/Bonsai/catalog/protocols/)',
    '  </Card>',
    '  <Card title="Sensors" icon="seti:config">',
    `    **${sensors.length} sensors** — Automated hook scripts that enforce boundaries, inject context, review output, and monitor code quality.`,
    '',
    '    [Browse sensors](/Bonsai/catalog/sensors/)',
    '  </Card>',
    '  <Card title="Routines" icon="seti:clock">',
    `    **${routines.length} routines** — Periodic self-maintenance tasks. Backlog hygiene, dependency audits, doc freshness checks, vulnerability scans, and more.`,
    '',
    '    [Browse routines](/Bonsai/catalog/routines/)',
    '  </Card>',
    '</CardGrid>',
  ];
  const changed = replaceMarkerContent(
    filePath,
    '{/* CATALOG-COUNTS-START */}',
    '{/* CATALOG-COUNTS-END */}',
    countLines.join('\n')
  );
  console.log(`overview.mdx: ${changed ? 'updated' : 'no changes'}`);
}

// ---------------------------------------------------------------------------
// Generate: catalog.json
// ---------------------------------------------------------------------------

{
  const catalogJson = {
    generated: new Date().toISOString(),
    counts: {
      agents: agents.length,
      skills: skills.length,
      workflows: workflows.length,
      protocols: protocols.length,
      sensors: sensors.length,
      routines: routines.length,
    },
    agents: agents.map((a) => ({
      name: a.name,
      display_name: a.display_name || displayName(a.name),
      description: a.description,
      defaults: a.defaults || {},
    })),
    skills: skills.map((s) => ({
      name: s.name,
      description: s.description,
      agents: s.agents,
      ...(s.triggers ? { triggers: s.triggers } : {}),
    })),
    workflows: workflows.map((w) => ({
      name: w.name,
      description: w.description,
      agents: w.agents,
      ...(w.triggers ? { triggers: w.triggers } : {}),
    })),
    protocols: protocols.map((p) => ({
      name: p.name,
      description: p.description,
      agents: p.agents,
      required: p.required || null,
    })),
    sensors: sensors.map((s) => ({
      name: s.name,
      description: s.description,
      agents: s.agents,
      event: s.event,
      matcher: s.matcher || null,
      required: s.required || null,
    })),
    routines: routines.map((r) => ({
      name: r.name,
      description: r.description,
      agents: r.agents,
      frequency: r.frequency,
    })),
  };

  const outPath = path.join(PUBLIC_DIR, 'catalog.json');
  fs.writeFileSync(outPath, JSON.stringify(catalogJson, null, 2) + '\n', 'utf8');
  console.log(`catalog.json: written to ${path.relative(WEBSITE_DIR, outPath)}`);
}

console.log('\nDone.');
