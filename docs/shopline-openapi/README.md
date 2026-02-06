# Shopline Open API Docs Mirror (Local)

This folder holds URL lists and (optionally) a locally mirrored plaintext/markdown copy of Shopline's Open API documentation, fetched via the `firecrawl` CLI.

## What’s Tracked

- `docs/shopline-openapi/urls.txt`: all discovered docs URLs from the reference sidebar.
- `docs/shopline-openapi/urls_endpoints.txt`: endpoint pages only (the ones we use for API coverage).
- `docs/shopline-openapi/urls_non_endpoints.txt`: category + index pages.

The bulk downloaded corpus is intentionally **not committed** (it’s large and changes often).

## Download / Refresh

Download endpoint docs as markdown + raw JSON responses:

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
