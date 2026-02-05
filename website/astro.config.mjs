import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://openboundary.org',
  prefetch: false,
  integrations: [
    starlight({
      title: 'OpenBoundary',
      description: 'Compile YAML specifications into type-safe TypeScript backends with enforced architectural constraints.',
      logo: {
        src: './src/assets/logo.svg',
      },
      social: {
        github: 'https://github.com/openboundary/openboundary',
      },
      customCss: [
        '@fontsource/ibm-plex-sans/400.css',
        '@fontsource/ibm-plex-sans/500.css',
        '@fontsource/ibm-plex-sans/600.css',
        '@fontsource/jetbrains-mono/400.css',
        '@fontsource/jetbrains-mono/500.css',
        '@fontsource/jetbrains-mono/600.css',
        '@fontsource/ibm-plex-mono/400.css',
        '@fontsource/ibm-plex-mono/500.css',
        './src/styles/global.css',
      ],
      head: [
        {
          tag: 'meta',
          attrs: {
            property: 'og:image',
            content: 'https://openboundary.org/images/og-image.png',
          },
        },
        {
          tag: 'meta',
          attrs: {
            name: 'twitter:card',
            content: 'summary_large_image',
          },
        },
      ],
      sidebar: [
        {
          label: 'Introduction',
          slug: 'docs',
        },
        {
          label: 'Quick Start',
          slug: 'docs/getting-started',
        },
        {
          label: 'Guides',
          items: [
            { label: 'Templates', slug: 'docs/templates' },
            { label: 'Claude Code Agent', slug: 'docs/agents' },
          ],
        },
        {
          label: 'Components',
          items: [
            { label: 'Overview', slug: 'docs/components' },
            { label: 'HTTP Server', slug: 'docs/components/http-server' },
            { label: 'PostgreSQL', slug: 'docs/components/postgres' },
            { label: 'Middleware', slug: 'docs/components/middleware' },
            { label: 'Use Case', slug: 'docs/components/usecase' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI', slug: 'docs/reference/cli' },
            { label: 'Specification Schema', slug: 'docs/reference/schema' },
            { label: 'Troubleshooting', slug: 'docs/reference/troubleshooting' },
          ],
        },
        {
          label: 'Roadmap',
          items: [
            { label: 'Architectures', slug: 'docs/architectures', badge: { text: 'Soon', variant: 'caution' } },
            { label: 'Changelog', slug: 'docs/roadmap' },
          ],
        },
      ],
      components: {
        // Using Starlight's default header for stability
        Footer: './src/components/Footer.astro',
      },
      expressiveCode: {
        themes: ['github-dark'],
        styleOverrides: {
          borderRadius: '8px',
          codeFontFamily: "'Geist Mono Variable', 'SF Mono', Consolas, monospace",
        },
      },
    }),
  ],
});
