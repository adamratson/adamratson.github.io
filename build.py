"""Build _site/index.html: a plain list of live GitHub Pages sites for a user."""

import html
import json
import os
import pathlib
import urllib.error
import urllib.request

USER = os.environ.get("GH_USER", "adamratson")
SELF = f"{USER}.github.io".lower()


def get_json(url):
    req = urllib.request.Request(url, headers={"Accept": "application/vnd.github+json"})
    token = os.environ.get("GITHUB_TOKEN")
    if token:
        req.add_header("Authorization", f"Bearer {token}")
    with urllib.request.urlopen(req) as resp:
        return json.load(resp)


def is_live(url):
    try:
        with urllib.request.urlopen(urllib.request.Request(url, method="HEAD")) as resp:
            return resp.status == 200
    except urllib.error.URLError:
        return False


repos = get_json(f"https://api.github.com/users/{USER}/repos?per_page=100&type=owner&sort=updated")
sites = []
for r in repos:
    if not r["has_pages"] or r["private"] or r["name"].lower() == SELF:
        continue
    url = f"https://{USER}.github.io/{r['name']}/"
    if is_live(url):
        sites.append((r["name"], url))

items = "\n".join(f'<li><a href="{url}">{html.escape(name)}</a></li>' for name, url in sites)
page = f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{USER}</title>
</head>
<body>
<h1>{USER}</h1>
<ul>
{items}
</ul>
</body>
</html>
"""

out = pathlib.Path("_site")
out.mkdir(exist_ok=True)
(out / "index.html").write_text(page)
print(f"Wrote {len(sites)} sites to _site/index.html")
