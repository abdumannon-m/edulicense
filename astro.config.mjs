// @ts-check
import { defineConfig } from 'astro/config';

import sitemap from '@astrojs/sitemap';

// Set PUBLIC_SITE_URL on Vercel (or change the fallback) so canonicals and sitemap URLs are correct.
const site =
	process.env.PUBLIC_SITE_URL?.replace(/\/$/, '') ??
	'https://edu-license.vercel.app';

// https://astro.build/config
export default defineConfig({
	site,
	trailingSlash: 'never',
	integrations: [sitemap()],
});
