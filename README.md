# Edu License LLC — landing site

Static marketing site for a Tashkent-based consultancy supporting schools in Uzbekistan with international licensing, institutional setup, and operational readiness.

- **Stack:** [Astro](https://astro.build/) (static output)
- **Locales:** English at `/`, Uzbek (Latin) at `/uz`
- **Conversion:** WhatsApp primary CTA ([`src/i18n/contact.ts`](src/i18n/contact.ts)); optional Telegram when `TELEGRAM_USERNAME` is set

## Local development

```sh
npm install
npm run dev
```

Open `http://localhost:4321` (English) and `http://localhost:4321/uz` (Uzbek).

```sh
npm run build    # output in dist/
npm run preview  # serve dist locally
npm run check    # astro check (TypeScript)
```

## Configuration

| Item | Where to change |
|------|------------------|
| Production URL (canonicals, sitemap) | `PUBLIC_SITE_URL` env var, or fallback in [`astro.config.mjs`](astro.config.mjs) |
| WhatsApp number, Telegram | [`src/i18n/contact.ts`](src/i18n/contact.ts) |
| All copy (EN / UZ) | [`src/i18n/en.ts`](src/i18n/en.ts), [`src/i18n/uz.ts`](src/i18n/uz.ts) |
| `robots.txt` sitemap host | [`public/robots.txt`](public/robots.txt) — keep in sync with your public domain |

## Deploying on Vercel

1. Push this repository to GitHub (or GitLab / Bitbucket).
2. In [Vercel](https://vercel.com/new), **Import** the repository.
3. Framework preset: **Astro** (or “Other” with build `npm run build`, output `dist`).
4. **Environment variables** (Production — and Preview if you want accurate URLs there):

   - `PUBLIC_SITE_URL` — full origin with no trailing slash, e.g. `https://your-domain.uz`

   Astro reads this in [`astro.config.mjs`](astro.config.mjs) for `site`, which drives canonical URLs and the sitemap.

5. Deploy. After the first deployment, update [`public/robots.txt`](public/robots.txt) so the `Sitemap:` line uses the same host as production (or add a later automation if you prefer).

**Custom domain:** In the Vercel project → **Settings → Domains**, add your domain and follow DNS instructions. Set `PUBLIC_SITE_URL` to that domain and redeploy so metadata and `sitemap-*.xml` use the correct origin.

## SEO notes

- [`src/layouts/BaseLayout.astro`](src/layouts/BaseLayout.astro) emits `canonical`, `hreflang` (`en`, `uz`, `x-default`), and basic Open Graph tags.
- `@astrojs/sitemap` generates `sitemap-index.xml` in `dist/` on build.

## Project structure

```text
src/
  components/     # Hero, sections, header, footer, WhatsApp FAB
  i18n/             # en / uz copy, contact helpers, hreflang URL helper
  layouts/          # BaseLayout (head, fonts)
  pages/
    index.astro     # English home
    uz/index.astro  # Uzbek home
    privacy.astro   # EN privacy stub
    uz/privacy.astro
  styles/global.css
public/
  favicon.svg
  robots.txt
```
