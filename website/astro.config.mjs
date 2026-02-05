import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://openboundary.org',
  integrations: [
    starlight({
      title: 'OpenBoundary',
      description: 'Define architectural and security invariants that AI agents can\'t violate. Generate type-safe backends from specifications.',
      logo: {
        src: './src/assets/logo.svg',
      },
      social: {
        github: 'https://github.com/openboundary/openboundary',
      },
      customCss: [
        '@fontsource/inter/400.css',
        '@fontsource/inter/500.css',
        '@fontsource/inter/600.css',
        '@fontsource-variable/geist-mono',
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
          label: 'Overview',
          slug: 'docs',
        },
        {
          label: 'Getting Started',
          slug: 'docs/getting-started',
        },
        {
          label: 'Components',
          items: [
            { label: 'Catalog', slug: 'docs/components' },
            { label: 'HTTP Server', slug: 'docs/components/http-server' },
            { label: 'PostgreSQL', slug: 'docs/components/postgres' },
            { label: 'Middleware', slug: 'docs/components/middleware' },
            { label: 'Use Case', slug: 'docs/components/usecase' },
          ],
        },
        {
          label: 'CLI Reference',
          slug: 'docs/cli',
        },
        {
          label: 'Roadmap',
          slug: 'docs/roadmap',
        },
      ],
      components: {
        Header: './src/components/Header.astro',
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
