// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// Deployed to GitHub Pages by .github/workflows/docs-release.yml.
// Sidebar entries are added by the milestone that delivers each feature
// (see the Development workflow guide in .okfw/).
// https://astro.build/config
export default defineConfig({
	site: 'https://infrashift.github.io',
	base: '/whac-a-leak',
	integrations: [
		starlight({
			title: 'Whac-a-Leak',
			description:
				'A lightweight, pure-Go secret scanner with self-testing CUE rules, layered org policy and a two-tier Aho-Corasick engine.',
			favicon: '/favicon.svg',
			customCss: ['./src/styles/custom.css'],
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/infrashift/whac-a-leak',
				},
			],
			editLink: {
				baseUrl: 'https://github.com/infrashift/whac-a-leak/edit/main/docs/',
			},
			sidebar: [
				{
					label: 'Start Here',
					items: [
						{ label: 'Overview', slug: 'docs' },
						{ label: 'Getting Started', slug: 'docs/getting-started' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'How Scanning Works', slug: 'docs/guides/how-scanning-works' },
						{ label: 'Writing Rules', slug: 'docs/guides/writing-rules' },
						{ label: 'Layering Org Policy', slug: 'docs/guides/layering' },
						{ label: 'Structured Files', slug: 'docs/guides/structured-files' },
						{ label: 'Allowlists & Fingerprints', slug: 'docs/guides/allowlists' },
						{ label: 'Pre-commit Hook', slug: 'docs/guides/pre-commit' },
						{ label: 'Scanning Git History', slug: 'docs/guides/git-history' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI', slug: 'docs/reference/cli' },
						{ label: 'Rule Schema', slug: 'docs/reference/rule-schema' },
						{ label: 'Output Formats', slug: 'docs/reference/output-formats' },
						{ label: 'Make Targets', slug: 'docs/reference/make-targets' },
					],
				},
				{
					label: 'Project',
					collapsed: true,
					items: [
						{ label: 'Roadmap', slug: 'docs/project/roadmap' },
						{ label: 'Design Decisions', slug: 'docs/project/decisions' },
						{ label: 'Contributing', slug: 'docs/project/contributing' },
					],
				},
			],
		}),
	],
});
