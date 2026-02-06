# Shopline Open API Docs Mirror (Local)

This folder holds URL lists and (optionally) a locally mirrored plaintext copy of Shopline's Open API documentation.

## What’s Tracked

- `docs/shopline-openapi/urls.txt`: all discovered docs URLs from the reference sidebar.
- `docs/shopline-openapi/urls_endpoints.txt`: endpoint pages only (the ones we use for API coverage).
- `docs/shopline-openapi/urls_non_endpoints.txt`: category + index pages.

The bulk downloaded corpus is intentionally **not committed** (it’s large and changes often).

## Download / Refresh (Preferred: Official `*.md`)

ReadMe exposes an official plaintext export for each API reference page:

- `https://open-api.docs.shoplineapp.com/reference/<slug>.md`

Download endpoint pages as plaintext markdown (local-only):

```bash
./scripts/download_shopline_reference_md.py --urls docs/shopline-openapi/urls_endpoints.txt
```

Outputs go to:

- `docs/shopline-openapi/pages_md/reference/*.md`

## Download / Refresh (Optional: Firecrawl)

If you want the browser-rendered reference pages scraped (e.g. for non-reference pages or troubleshooting), you can still use Firecrawl.

Download endpoint docs as Firecrawl markdown + raw JSON responses:

```bash
./scripts/download_shopline_openapi_docs.py --urls docs/shopline-openapi/urls_endpoints.txt
```

By default this uses:

- `onlyMainContent=false` (the endpoint URL + method often disappear when `onlyMainContent=true`)
- `excludeTags=script,style` (prevents huge pages / corrupted cached outputs)
- automatic fallback to `--max-age 0` when cached scrapes are invalid JSON

Download non-endpoint pages (optional):

```bash
./scripts/download_shopline_openapi_docs.py --urls docs/shopline-openapi/urls_non_endpoints.txt --jobs 1
```

Outputs go to:

- `docs/shopline-openapi/pages/**/*.md`
- `docs/shopline-openapi/pages/**/*.json`

## Notes

- Some non-endpoint pages can be extremely large; scraping may time out.
- The endpoint pages are the primary source of truth for CLI coverage planning.
