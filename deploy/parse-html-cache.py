#!/usr/bin/env python3
"""Parse cached HTML files into chunks JSON, one file at a time.
Writes NDJSON then converts. Minimal memory usage."""
import gc
import json
import os
import re
import sys
from html.parser import HTMLParser


class CE(HTMLParser):
    def __init__(self):
        super().__init__()
        self.title = ""
        self.parts = []
        self._it = self._im = self._ib = False
        self._sd = 0

    _skip = frozenset(("script", "style", "nav", "footer", "svg", "noscript"))

    def handle_starttag(self, tag, attrs):
        if tag in self._skip:
            self._sd += 1
        elif tag == "title":
            self._it = True
        elif tag in ("main", "article"):
            self._im = True
        elif tag == "body":
            self._ib = True

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


def main():
    htmldir = sys.argv[1]
    output = sys.argv[2]

    url_map = {}
    with open(os.path.join(htmldir, "url-map.txt")) as f:
        for line in f:
            parts = line.strip().split(" ", 1)
            if len(parts) == 2:
                url_map[parts[0]] = parts[1]

    ndjson = output + ".ndjson"
    total = 0
    parsed = 0

    # Write one chunk per line — no list accumulation
    with open(ndjson, "w") as out:
        for safename, url in sorted(url_map.items()):
            htmlfile = os.path.join(htmldir, safename + ".html")
            if not os.path.exists(htmlfile):
                continue

            with open(htmlfile, "r", errors="replace") as hf:
                html = hf.read(262144)

            p = CE()
            p.feed(html)
            content = re.sub(r"\s+", " ", " ".join(p.parts)).strip()[:8000]
            title = p.title.strip() if p.title.strip() else url
            del html, p

            if not content:
                continue

            start = 0
            idx = 0
            while start < len(content):
                end = min(start + 2048, len(content))
                if end < len(content):
                    for delim in [". ", ".\n", "! ", "? "]:
                        pos = content.rfind(delim, start, end)
                        if pos > start:
                            end = pos + 1
                            break
                chunk = content[start:end].strip()
                if chunk:
                    out.write(json.dumps({"text": chunk, "url": url, "title": title, "index": idx}))
                    out.write("\n")
                    idx += 1
                    total += 1
                # Advance by at least 1 char to prevent infinite loop
                next_start = max(end - 200, start + 1)
                if next_start >= len(content):
                    break
                start = next_start

            del content
            parsed += 1
            if parsed % 10 == 0:
                gc.collect()
                out.flush()

    print(f"Parsed {parsed} pages -> {total} chunks", file=sys.stderr)

    # Convert NDJSON to JSON array (streaming)
    with open(output, "w") as out:
        out.write("[")
        first = True
        with open(ndjson) as nd:
            for line in nd:
                line = line.strip()
                if not line:
                    continue
                if not first:
                    out.write(",")
                out.write(line)
                first = False
        out.write("]")

    os.remove(ndjson)
    print(f"Saved to {output}", file=sys.stderr)


if __name__ == "__main__":
    main()
