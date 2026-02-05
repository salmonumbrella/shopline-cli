#!/usr/bin/env python3
"""Download Shopline Open API docs pages as markdown using firecrawl CLI.

Inputs:
  - docs/shopline-openapi/urls.txt (one URL per line)

Outputs:
  - docs/shopline-openapi/pages/<path>.md
  - docs/shopline-openapi/pages/<path>.json (raw firecrawl response)

This script is resumable: it skips pages that already have a .md file.
"""

import argparse
import concurrent.futures
import json
import os
import re
import subprocess
import sys
import threading
from pathlib import Path


def safe_rel_path(url: str) -> str:
    # Keep domain path, strip scheme+host.
    m = re.match(r"^https?://[^/]+(/.*)$", url)
    path = m.group(1) if m else url
    path = path.split("#", 1)[0]
    path = path.split("?", 1)[0]
    # Normalize.
    path = path.strip("/")
    if not path:
        return "root"
    # Avoid weird filesystem chars.
    path = re.sub(r"[^A-Za-z0-9._\-/]", "_", path)
    return path


def run_firecrawl_scrape(url: str, max_age_ms: int) -> dict:
    raw = {
        "url": url,
        "formats": ["markdown"],
        "onlyMainContent": True,
        "maxAge": max_age_ms,
    }
    cmd = [
        "firecrawl",
        "scrape",
        "--url",
        url,
        "--formats",
        "markdown",
        "--only-main-content",
        "true",
        "--raw",
        json.dumps(raw),
    ]
    p = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if p.returncode != 0:
        raise RuntimeError(f"firecrawl scrape failed (rc={p.returncode}): {p.stderr.strip()}")
    try:
        return json.loads(p.stdout)
    except json.JSONDecodeError as e:
        raise RuntimeError(f"firecrawl returned non-JSON for {url}: {e}\nstdout prefix: {p.stdout[:200]}")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--urls", default="docs/shopline-openapi/urls.txt")
    ap.add_argument("--out", default="docs/shopline-openapi/pages")
    ap.add_argument("--max-age-ms", type=int, default=172800000)  # 48h
    ap.add_argument("--jobs", type=int, default=4)
    ap.add_argument("--limit", type=int, default=0, help="0 = no limit")
    ap.add_argument("--force", action="store_true", help="re-download even if md exists")
    args = ap.parse_args()

    urls_path = Path(args.urls)
    out_dir = Path(args.out)
    out_dir.mkdir(parents=True, exist_ok=True)

    urls = [
        line.strip()
        for line in urls_path.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.strip().startswith("#")
    ]

    total = 0
    skipped = 0
    failed = 0

    lock = threading.Lock()

    def one(url: str) -> None:
        nonlocal total, skipped, failed

        rel = safe_rel_path(url)
        md_path = out_dir / (rel + ".md")
        raw_path = out_dir / (rel + ".json")
        md_path.parent.mkdir(parents=True, exist_ok=True)

        if md_path.exists() and not args.force:
            with lock:
                skipped += 1
            return

        try:
            data = run_firecrawl_scrape(url, args.max_age_ms)
            raw_path.write_text(json.dumps(data, indent=2, ensure_ascii=False), encoding="utf-8")
            md = data.get("markdown")
            if not isinstance(md, str):
                raise RuntimeError(f"missing markdown field (keys={list(data.keys())})")
            md_path.write_text(md, encoding="utf-8")
            with lock:
                total += 1
        except Exception as e:
            with lock:
                failed += 1
            print(f"[FAIL] {url}: {e}", file=sys.stderr)

    # Apply --limit to the processing queue (downloaded+failed), not the input list size.
    queue = urls if not args.limit else urls[: args.limit]
    jobs = max(1, int(args.jobs))
    with concurrent.futures.ThreadPoolExecutor(max_workers=jobs) as ex:
        list(ex.map(one, queue))

    print(json.dumps({"downloaded": total, "skipped": skipped, "failed": failed, "total_urls": len(urls)}))
    return 0 if failed == 0 else 2


if __name__ == "__main__":
    raise SystemExit(main())
