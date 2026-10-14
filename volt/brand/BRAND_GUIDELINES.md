# Volt API — brand guidelines & handover

The brand decision record for **Volt API**, so every repo in the family (this Go backend, plus **Volt Web** and any future surface) uses the *same* name, colors, logo, and typography. This is the source of truth for the identity; the concrete assets and generator live alongside it in this `brand/` directory. It is intentionally near-identical to the Volt Web repo's copy — the brand is one brand across the family.

## 1. Name & family

- **Product name:** **Volt API** (this repo — the Go backend starter). The umbrella brand is **Volt**; lowercase `volt` as a slug/package id, `volt-api` for this surface.
- **Concept:** a lightning bolt — "Volt = powered-up, fast to start." The name reinforces the mark and vice-versa.
- **Handle (technical identity):** **`usevolt`** — use this wherever a unique string is needed: domain `usevolt.dev`, GitHub org `usevolt`, npm scope `@usevolt/*`.
- **Family / naming convention:** Volt is an umbrella over multiple starters. Suffix the surface:
  - **Volt Web** — the TanStack Start React frontend. Package `volt-web` / `@usevolt/web`.
  - **Volt API** — this Go backend starter. Package/repo `volt-api` / `@usevolt/api`.
  - Extend the same way for anything else (`Volt CLI`, `Volt UI`, …).

## 2. Taglines

- **Volt API:** "The batteries-included Go backend starter."
- **Volt Web (parallel):** "The batteries-included TanStack Start starter." — keep the "batteries-included … starter" cadence across the family.
- **OG / social subline pattern:** short feature list separated by `·`, e.g. `GraphQL · Ent ORM · NATS · typed config · codegen`.

## 3. Logo & mark

**The mark is a lightning bolt on a rounded, violet gradient tile.** It is the only per-product art — everything else (color tiles, favicon, wordmark, lockups, OG) reflows from it.

Geometry (so every surface stays identical):

- **512×512 master**, tile is a `rect` with corner radius **112** (≈ 22%).
- The bolt glyph, drawn in the 0–512 box, is a single path — reuse it verbatim:
  ```
  M286 96 L166 280 H240 L226 416 L346 232 H272 L286 96 Z
  ```
  Filled + stroked with the foreground color, `stroke-width="8"`, `stroke-linejoin="round"`.
- **Variants** swap the tile + foreground, never the geometry: color tile → white bolt; dark tile → light bolt; light tile → violet bolt; glyph-only (no tile) → violet bolt; mono → `currentColor`.
- **Favicon** uses the same bolt (legible at 16px).

Don'ts: don't recolor outside the palette, don't rotate/stretch/skew the mark, don't add shadows or bevels (it's flat), keep clearspace ≥ 25% of the icon width.

## 4. Color

Violet gradient identity (also machine-readable in `palette.json` / `tokens.json` / `tokens.css`):

| Role | Value(s) | Use |
|---|---|---|
| **Tile gradient** | `#6366F1` → `#8B5CF6` | primary icon tile (indigo → violet) |
| **Accent** | `#7C3AED` | wordmark accent, links, highlights, OG accent line |
| **Spark / light accent** | `#A78BFA` | subtle highlights on the color tile |
| **Foreground on tile** | `#FFFFFF` | the bolt on the color tile |
| **Dark tile** | `#4C1D95` → `#2E1065` | tile variant for bright surfaces; dark-mode backgrounds |
| **Dark foreground** | `#EDE9FE` | glyph/text on dark tiles |
| **Light tile** | `#EDE9FE` → `#DDD6FE` | tile variant for dark surfaces |

Keep these hex values identical across the family. If a surface needs a per-service tint, add a *tile variant* in the same violet family rather than a new hue.

## 5. Typography

- **Display / wordmark:** **Space Grotesk**, with an **Inter → system-ui** fallback stack (the shipped raster assets were generated with the Inter fallback).
- **Sans (body/UI):** **Inter**.
- **Mono:** **JetBrains Mono**.

All three are open-source (OFL), free to self-host. These are the *brand* fonts (wordmark, marketing, OG); the backend service itself has no UI. Full stacks live in `tokens.css`.

## 6. The asset kit (this directory)

Everything is generated from one script so a color/name change reflows the whole set. **Never hand-edit the output SVG/PNG** — edit the generator or the manifest and regenerate.

| File | What |
|---|---|
| `brand.json` | the manifest (name `volt-api`, tagline, colors, fonts, topics, positioning) |
| `gen-brand.mjs` | the generator — the single source of truth for all art |
| `Taskfile.yml` | `task gen` / `raster` / `all` / `clean` |
| `icon.svg`, `volt-icon.svg`, `volt-icon-dark.svg`, `volt-icon-light.svg` | the color-tile icon + variants |
| `volt-glyph.svg`, `volt-mono.svg` | bolt only / single-color |
| `favicon.svg`, `favicon.ico` | favicon (multi-res `.ico`: 16/32/48) |
| `volt-wordmark.svg`, `volt-lockup.svg`, `volt-lockup-dark.svg` | wordmark + horizontal lockups |
| `og-cover.svg`, `og-cover-light.svg` | 1200×630 social cards (dark + light) |
| `palette.json`, `tokens.json`, `tokens.css` | design tokens |
| `png/` | rasters incl. `og-social-1280.png` (GitHub social preview, 1280×640) |
| `README.md` | brand usage quick-reference |

Regenerate everything (needs node + rsvg-convert + ImageMagick):

```bash
task -t volt/brand/Taskfile.yml all
```

## 7. Reusing Volt branding in another repo (e.g. Volt Web)

1. **Copy the `brand/` directory** into the other repo.
2. In `brand/brand.json`, change `name` (e.g. `"volt-web"`), `repo`, `language`, `kind`, and the tagline/description/topics for that surface. **Keep the `brand` (colors) and `fonts` blocks identical.**
3. In `brand/gen-brand.mjs`, keep the manifest colors/fonts and the **bolt `GLYPH()` body identical**; only the OG tagline/subline copy differs per surface.
4. Run the generator to emit that repo's own icon/favicon/OG set in the same identity.

## 8. Summary of decisions

- Product = **Volt API** (Go backend surface of **Volt**); mark = **lightning bolt**, palette = **violet gradient** (`#6366F1`→`#8B5CF6`, accent `#7C3AED`).
- Technical handle = **`usevolt`** (`usevolt.dev`, `@usevolt/*`, GitHub org `usevolt`); family naming = **Volt Web** / **Volt API** (`volt-web` / `volt-api`).
- Fonts = **Space Grotesk** (display) / **Inter** (sans) / **JetBrains Mono** (mono).
- Full regenerable brand kit lives in this `brand/`, driven by `brand.json` + `gen-brand.mjs`.
- Distribution = **open source, Apache-2.0** — see the repo `LICENSE`.
