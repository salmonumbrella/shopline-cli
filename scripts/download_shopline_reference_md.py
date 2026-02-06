#!/usr/bin/env python3
"""Download Shopline API reference pages via the official ReadMe plaintext route.

Shopline's ReadMe-hosted API reference supports a plaintext endpoint for each
reference page:

  https://open-api.docs.shoplineapp.com/reference/<slug>.md

This is often more reliable than browser rendering or third-party scraping and
it embeds the endpoint's OpenAPI snippet, which we use for coverage indexing.

Inputs:
  - docs/shopline-openapi/urls_endpoints.txt (one /reference/... URL per line)

Outputs (local-only; intentionally not committed):
  - docs/shopline-openapi/pages_md/<path>.md
    Example: docs/shopline-openapi/pages_md/reference/get_orders-1.md
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import random
import re
import sys
import threading
import time
import urllib.error
import urllib.request
from pathlib import Path


def safe_rel_path(url: str) -> str:
    # Keep domain path, strip scheme+host.
    m = re.match(r"^https?://[^/]+(/.*)$", url)
    path = m.group(1) if m else url
    path = path.split("#", 1)[0]
    path = path.split("?", 1)[0]
    path = path.strip("/")
    if not path:
        return "root"
    # Keep "/" to preserve directory structure.
    return re.sub(r"[^A-Za-z0-9._/-]", "_", path)


def to_md_url(u: str) -> str:
    u = u.strip()
    if not u:
        return u
    if u.endswith(".md"):
        return u
    return u + ".md"


def fetch_text(url: str, timeout_s: int) -> tuple[int, str]:
    req = urllib.request.Request(
        url,
        headers={
            "User-Agent": "shopline-cli-docs-mirror/1.0",
            "Accept": "text/plain, text/markdown;q=0.9, */*;q=0.1",
        },
        method="GET",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout_s) as resp:
            status = int(getattr(resp, "status", 200))
            data = resp.read()
            text = data.decode("utf-8", errors="replace")
            return status, text
    except urllib.error.HTTPError as e:
        body = ""
        try:
            body = e.read().decode("utf-8", errors="replace")
        except Exception:
            body = ""
        return int(getattr(e, "code", 0) or 0), body


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--urls", default="docs/shopline-openapi/urls_endpoints.txt")
    ap.add_argument("--out", default="docs/shopline-openapi/pages_md")
    ap.add_argument("--jobs", type=int, default=6)
    ap.add_argument("--timeout", type=int, default=30)
    ap.add_argument("--retries", type=int, default=3)
    ap.add_argument("--force", action="store_true", help="re-download even if the md file exists")
    ap.add_argument("--limit", type=int, default=0, help="0 = no limit")
    args = ap.parse_args()

    urls_path = Path(args.urls)
    out_dir = Path(args.out)
    out_dir.mkdir(parents=True, exist_ok=True)

    urls = [
        line.strip()
        for line in urls_path.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.strip().startswith("#")
    ]
    if args.limit and args.limit > 0:
        urls = urls[: args.limit]

    downloaded = 0
    skipped = 0
    failed = 0
    lock = threading.Lock()

    def one(url: str) -> None:
        nonlocal downloaded, skipped, failed

        md_url = to_md_url(url)
        rel = safe_rel_path(md_url)
        md_path = out_dir / rel
        if md_path.suffix != ".md":
            md_path = out_dir / (rel + ".md")
        md_path.parent.mkdir(parents=True, exist_ok=True)

        if md_path.exists() and not args.force:
            with lock:
                skipped += 1
            return

        last = None
        for attempt in range(int(args.retries) + 1):
            status, text = fetch_text(md_url, timeout_s=int(args.timeout))
            if status == 200 and text:
                md_path.write_text(text, encoding="utf-8")
                with lock:
                    downloaded += 1
                return

            # Back off on throttling / transient errors.
            if status in (429, 500, 502, 503, 504):
                # Jittered exponential backoff: ~0.5s, 1s, 2s, ...
                delay = (0.5 * (2**attempt)) + random.random() * 0.2
                time.sleep(delay)
                last = f"status={status}"
                continue

            last = f"status={status}"
            break

        with lock:
            failed += 1
        print(f"[FAIL] {md_url}: {last}", file=sys.stderr)

    jobs = max(1, int(args.jobs))
    with concurrent.futures.ThreadPoolExecutor(max_workers=jobs) as ex:
        list(ex.map(one, urls))

    print(json.dumps({"downloaded": downloaded, "skipped": skipped, "failed": failed, "total_urls": len(urls)}))
    return 0 if failed == 0 else 2


if __name__ == "__main__":
    raise SystemExit(main())
