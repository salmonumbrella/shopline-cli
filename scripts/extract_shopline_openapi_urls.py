#!/usr/bin/env python3
"""Extract Shopline Open API docs URLs from the reference sidebar.

This script uses `firecrawl scrape` to fetch the reference HTML, then extracts
all `/reference/...` and `/docs...` links.

Outputs:
  - docs/shopline-openapi/urls.txt
  - docs/shopline-openapi/urls_endpoints.txt
  - docs/shopline-openapi/urls_non_endpoints.txt

Note: The bulk downloaded corpus is handled by `scripts/download_shopline_openapi_docs.py`.
"""

import argparse
import json
import re
import subprocess
from pathlib import Path
from urllib.parse import urlparse


def firecrawl_scrape_html(url: str, max_age_ms: int) -> str:
    raw = {"url": url, "formats": ["html"], "onlyMainContent": False, "maxAge": max_age_ms}
    cmd = [
        "firecrawl",
        "scrape",
        "--url",
        url,
        "--formats",
        "html",
        "--only-main-content",
        "false",
        "--raw",
        json.dumps(raw),
    ]
    p = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if p.returncode != 0:
        raise SystemExit(f"firecrawl scrape failed (rc={p.returncode}): {p.stderr.strip()}")
    data = json.loads(p.stdout)
    html = data.get("html")
    if not isinstance(html, str):
        raise SystemExit(f"unexpected firecrawl response keys={list(data.keys())}")
    return html


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "--url",
        default="https://open-api.docs.shoplineapp.com/reference",
        help="Reference page URL. (Prefer the reference index; some endpoint pages can be too large to scrape as HTML.)",
    )
    ap.add_argument("--out-dir", default="docs/shopline-openapi")
    ap.add_argument("--max-age-ms", type=int, default=172800000)
    args = ap.parse_args()

    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    # Cache the reference sidebar HTML locally to avoid repeatedly hitting firecrawl
    # (and to avoid occasional truncation on very large pages).
    discovery_dir = out_dir / "_discovery"
    discovery_dir.mkdir(parents=True, exist_ok=True)
    cached = discovery_dir / "reference.html.json"

    html = ""
    if cached.exists():
        try:
            data = json.loads(cached.read_text(encoding="utf-8"))
            if isinstance(data, dict) and isinstance(data.get("html"), str):
                html = data["html"]
        except Exception:
            html = ""

    if not html:
        raw = {"url": args.url, "formats": ["html"], "onlyMainContent": False, "maxAge": args.max_age_ms}
        cmd = [
            "firecrawl",
            "scrape",
            "--url",
            args.url,
            "--formats",
            "html",
            "--only-main-content",
            "false",
            "--raw",
            json.dumps(raw),
        ]
        p = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        if p.returncode != 0:
            raise SystemExit(f"firecrawl scrape failed (rc={p.returncode}): {p.stderr.strip()}")
        cached.write_text(p.stdout, encoding="utf-8")
        data = json.loads(p.stdout)
        html = data.get("html", "")

    hrefs = set(re.findall(r'href="([^"]+)"', html))

    urls = set()
    for h in hrefs:
        if h.startswith("https://open-api.docs.shoplineapp.com/"):
            urls.add(h.split("#", 1)[0])
        elif h.startswith("/"):
            urls.add("https://open-api.docs.shoplineapp.com" + h.split("#", 1)[0])

    keep = []
    for u in urls:
        if urlparse(u).netloc != "open-api.docs.shoplineapp.com":
            continue
        if u.startswith("https://open-api.docs.shoplineapp.com/reference") or u.startswith(
            "https://open-api.docs.shoplineapp.com/docs"
        ):
            keep.append(u)
        elif u in {
            "https://open-api.docs.shoplineapp.com/changelog",
            "https://open-api.docs.shoplineapp.com/discuss",
        }:
            keep.append(u)

    keep = sorted(set(keep))

    endpoints = []
    for u in keep:
        if not u.startswith("https://open-api.docs.shoplineapp.com/reference/"):
            continue
        slug = u.split("/reference/", 1)[1]
        if re.match(r"^(get|post|put|patch|delete|del)_[A-Za-z0-9].*", slug):
            endpoints.append(u)

    non_endpoints = sorted(set(keep) - set(endpoints))

    (out_dir / "urls.txt").write_text("\n".join(keep) + "\n", encoding="utf-8")
    (out_dir / "urls_endpoints.txt").write_text("\n".join(sorted(endpoints)) + "\n", encoding="utf-8")
    (out_dir / "urls_non_endpoints.txt").write_text("\n".join(non_endpoints) + "\n", encoding="utf-8")

    print(json.dumps({"all": len(keep), "endpoints": len(endpoints), "non_endpoints": len(non_endpoints)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
