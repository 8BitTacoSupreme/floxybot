#!/usr/bin/env python3
"""
Scrape flox.dev docs and blog, chunk content, save as JSON.
Memory-efficient: processes one page at a time, writes chunks incrementally.
Uses urllib (no subprocess fork overhead).
"""
import gc
import json
import os
import re
import ssl
import sys
import urllib.request
from collections import deque
from html.parser import HTMLParser
from urllib.parse import urljoin

CHUNK_SIZE = 2048
OVERLAP = 200
MAX_PAGES = 50
OUTPUT = os.environ.get("CANON_OUTPUT", "/opt/floxybot/data/canon/chunks.json")

BASE_URLS = [
    "https://flox.dev/docs/",
    "https://flox.dev/blog/",
]

# Reuse a single SSL context to avoid repeated handshake overhead.
_ssl_ctx = ssl.create_default_context()


class ContentExtractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.title = ""
        self.parts = []
        self.links = []
        self._it = self._im = self._ib = False
        self._sd = 0

    _skip = frozenset(("script", "style", "nav", "footer", "svg", "noscript"))

    def handle_starttag(self, tag, attrs):
        if tag in self._skip:
            self._sd += 1
            return
        if tag == "title":
            self._it = True
        elif tag in ("main", "article"):
            self._im = True
        elif tag == "body":
            self._ib = True
        elif tag == "a":
            for k, v in attrs:
                if k == "href" and v:
                    self.links.append(v)
                    break

    def handle_endtag(self, tag):
        if tag in self._skip and self._sd > 0:
            self._sd -= 1
        elif tag == "title":
            self._it = False

    def handle_data(self, data):
        if self._sd > 0:
            return
        text = data.strip()
        if not text:
            return
        if self._it:
            self.title += text
        if self._im or self._ib:
            self.parts.append(text)


def fetch(url):
    """Fetch URL, return HTML string or None."""
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "floxybot-scraper/1.0"})
        with urllib.request.urlopen(req, timeout=30, context=_ssl_ctx) as resp:
            data = resp.read(262144)
        return data.decode("utf-8", errors="replace")
    except Exception as e:
        print(f"  skip {url}: {e}", file=sys.stderr)
        return None


def process_page(html_text, base_url):
    """Parse HTML, return (title, content_text, discovered_links)."""
    p = ContentExtractor()
    p.feed(html_text)
    content = re.sub(r"\s+", " ", " ".join(p.parts)).strip()
    if len(content) > 8000:
        content = content[:8000]

    links = []
    seen = set()
    for href in p.links:
        if href.startswith(("#", "javascript:", "mailto:")):
            continue
        full = urljoin(base_url, href).split("#")[0]
        if "flox.dev" in full and full not in seen:
            seen.add(full)
            links.append(full)

    return p.title.strip(), content, links


def chunk_text(text, url, title):
    """Split text into overlapping chunks."""
    chunks = []
    start = 0
    idx = 0
    while start < len(text):
        end = min(start + CHUNK_SIZE, len(text))
        if end < len(text):
            for delim in [". ", ".\n", "! ", "? "]:
                pos = text.rfind(delim, start, end)
                if pos > start:
                    end = pos + 1
                    break
        chunk = text[start:end].strip()
        if chunk:
            chunks.append({"text": chunk, "url": url, "title": title, "index": idx})
            idx += 1
        next_start = max(end - OVERLAP, start + 1)
        if next_start >= len(text):
            break
        start = next_start
    return chunks


def main():
    os.makedirs(os.path.dirname(OUTPUT) or ".", exist_ok=True)

    visited = set()
    queue = deque(BASE_URLS)
    total_pages = 0
    total_chunks = 0

    # Write NDJSON (one chunk per line), convert to array at end
    ndjson_path = OUTPUT + ".tmp"

    with open(ndjson_path, "w") as ndf:
        while queue and total_pages < MAX_PAGES:
            url = queue.popleft()
            if url in visited:
                continue
            visited.add(url)

            html_text = fetch(url)
            if html_text is None:
                continue

            title, content, links = process_page(html_text, url)
            # Free HTML immediately
            del html_text

            if not content:
                del links
                gc.collect()
                continue

            chunks = chunk_text(content, url, title)
            del content

            for c in chunks:
                ndf.write(json.dumps(c))
                ndf.write("\n")
                total_chunks += 1
            del chunks

            total_pages += 1

            # Enqueue discovered links
            for link in links:
                if link not in visited and ("/docs/" in link or "/blog/" in link):
                    queue.append(link)
            del links

            if total_pages % 5 == 0:
                ndf.flush()
                gc.collect()

            print(f"  {total_pages}: {url} ({total_chunks} chunks)", file=sys.stderr)

    # Convert NDJSON to JSON array (streaming read to avoid loading all at once)
    print(f"Converting {total_chunks} chunks to JSON array...", file=sys.stderr)
    with open(OUTPUT, "w") as out:
        out.write("[")
        first = True
        with open(ndjson_path) as nd:
            for line in nd:
                line = line.strip()
                if not line:
                    continue
                if not first:
                    out.write(",")
                out.write(line)
                first = False
        out.write("]")

    os.remove(ndjson_path)
    print(f"Done: {total_chunks} chunks from {total_pages} pages -> {OUTPUT}", file=sys.stderr)


if __name__ == "__main__":
    main()
