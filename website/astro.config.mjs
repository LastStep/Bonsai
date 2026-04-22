import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightLlmsTxt from 'starlight-llms-txt';
import starlightLinksValidator from 'starlight-links-validator';

export default defineConfig({
  site: 'https://laststep.github.io/Bonsai',
  base: '/Bonsai',
  integrations: [
    starlight({
      title: 'Bonsai',
      description: 'A workspace for your coding agent',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/LastStep/Bonsai',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/LastStep/Bonsai/edit/main/website/',
      },
      lastUpdated: true,
      customCss: ['./src/styles/custom.css'],
      plugins: [
        starlightLlmsTxt({
          projectName: 'Bonsai',
          description: `Bonsai is a CLI tool for scaffolding Claude Code agent workspaces. It generates structured instruction files — identity, memory, protocols, skills, workflows, sensors, and routines — so AI agents work like teammates, not tools.`,
          details: `- Install: \`go install github.com/LastStep/Bonsai@latest\` or \`brew install LastStep/tap/bonsai\`
- 6 agent types: tech-lead, backend, frontend, fullstack, devops, security
- Abilities are modular: skills (reference), workflows (multi-step), protocols (rules), sensors (hooks), routines (periodic)`,
          customSets: [
            {
              label: 'Concepts',
              description: 'How Bonsai works — agents, abilities, sensors, routines, scaffolding, workspaces',
              paths: ['concepts/**'],
            },
            {
              label: 'Commands',
              description: 'CLI reference for all 7 commands with flags and examples',
              paths: ['commands/**'],
            },
            {
              label: 'Catalog',
              description: 'All 6 agent types, 17 skills, 10 workflows, 4 protocols, 12 sensors, 8 routines with descriptions and compatibility',
              paths: ['catalog/**'],
            },
            {
              label: 'Configuration',
              description: '.bonsai.yaml, .bonsai-lock.yaml, meta.yaml, and agent.yaml schemas',
              paths: ['reference/**'],
            },
          ],
        }),
        starlightLinksValidator(),
      ],
      sidebar: [
        {
          label: 'Start Here',
          items: [
            { slug: 'getting-started' },
            { slug: 'installation' },
            { slug: 'why-bonsai' },
          ],
        },
        {
          label: 'Core Concepts',
          items: [
            { slug: 'concepts/how-bonsai-works' },
            { slug: 'concepts/agents' },
            { slug: 'concepts/abilities' },
            { slug: 'concepts/sensors' },
            { slug: 'concepts/routines' },
            { slug: 'concepts/scaffolding' },
            { slug: 'concepts/workspaces' },
          ],
        },
        {
          label: 'Commands',
          items: [
            { slug: 'commands/init' },
            { slug: 'commands/add' },
            { slug: 'commands/remove' },
            { slug: 'commands/list' },
            { slug: 'commands/catalog' },
            { slug: 'commands/update' },
            { slug: 'commands/guide' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { slug: 'guides/your-first-workspace' },
            { slug: 'guides/working-with-agents' },
            { slug: 'guides/triggers-and-activation' },
            { slug: 'guides/customizing-abilities' },
            { slug: 'guides/creating-custom-skills' },
            { slug: 'guides/creating-custom-sensors' },
            { slug: 'guides/creating-custom-routines' },
            { slug: 'guides/dogfooding' },
          ],
        },
        {
          label: 'Catalog',
          collapsed: true,
          items: [
            { slug: 'catalog/overview' },
            { slug: 'catalog/agent-types' },
            { slug: 'catalog/skills' },
            { slug: 'catalog/workflows' },
            { slug: 'catalog/protocols' },
            { slug: 'catalog/sensors' },
            { slug: 'catalog/routines' },
          ],
        },
        {
          label: 'Reference',
          collapsed: true,
          items: [
            { slug: 'reference/configuration' },
            { slug: 'reference/lock-file' },
            { slug: 'reference/template-variables' },
            { slug: 'reference/meta-yaml-schema' },
            { slug: 'reference/agent-yaml-schema' },
            { slug: 'reference/glossary' },
          ],
        },
      ],
    }),
  ],
});
