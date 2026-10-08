#!/usr/bin/env python3
"""Build http/postman_collection.json from the kulala .http files in http/.

One folder per .http file, one request per "###" block. {{base}} and other
values from the first environment in http-client.env.json become collection
variables. Kulala response references like {{createHost.response.body.$.id}}
become {{createHost_id}}, set by a test script on the "# @name createHost" request.
"""
import json
import re
import sys
from pathlib import Path

HTTP_DIR = Path(sys.argv[1] if len(sys.argv) > 1 else "http")
OUT = HTTP_DIR / "postman_collection.json"

REF = re.compile(r"\{\{\s*(\w+)\.response\.body\.(\$[^}\s]*)\s*\}\}")
METHODS = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}


def var_name(req, path):
    return req + "_" + re.sub(r"\W+", "_", path.lstrip("$.")).strip("_")


def to_postman(text):
    return REF.sub(lambda m: "{{" + var_name(m[1], m[2]) + "}}", text)


def parse(path):
    blocks, cur = [], None
    for line in path.read_text().splitlines():
        if line.startswith("###"):
            cur = {"title": line[3:].strip(), "lines": []}
            blocks.append(cur)
        elif cur is None:
            cur = {"title": "", "lines": [line]}
            blocks.append(cur)
        else:
            cur["lines"].append(line)

    requests, file_vars = [], {}
    for b in blocks:
        name, method, url, headers, body = None, None, None, [], []
        stage = "pre"
        for line in b["lines"]:
            s = line.strip()
            if stage == "pre":
                if m := re.match(r"#\s*@name\s+(\S+)", s):
                    name = m[1]
                elif m := re.match(r"@(\w+)\s*=\s*(.*)", s):
                    file_vars[m[1]] = m[2]
                elif not s or s.startswith("#") or s.startswith("//"):
                    continue
                else:
                    parts = s.split()
                    if parts[0] in METHODS:
                        method, url = parts[0], parts[1]
                    else:
                        method, url = "GET", parts[0]
                    stage = "headers"
            elif stage == "headers":
                if not s:
                    stage = "body"
                elif not s.startswith("#"):
                    k, _, v = s.partition(":")
                    headers.append({"key": k.strip(), "value": v.strip()})
            else:
                if s.startswith(">") or s.startswith("<>"):
                    break
                body.append(line)
        if method:
            requests.append({
                "title": b["title"] or f"{method} {url}",
                "name": name,
                "method": method,
                "url": url,
                "headers": headers,
                "body": "\n".join(body).strip(),
            })
    return requests, file_vars


def main():
    files = sorted(HTTP_DIR.glob("*.http"))
    parsed = {f: parse(f) for f in files}

    variables = {}
    env_file = HTTP_DIR / "http-client.env.json"
    if env_file.exists():
        envs = json.loads(env_file.read_text())
        envs.pop("$shared", None)
        if envs:
            variables.update(next(iter(envs.values())))

    refs = {}
    for reqs, file_vars in parsed.values():
        variables.update(file_vars)
        for r in reqs:
            for text in [r["url"], r["body"], *(h["value"] for h in r["headers"])]:
                for m in REF.finditer(text):
                    refs.setdefault(m[1], {})[var_name(m[1], m[2])] = m[2]

    folders = []
    for f, (reqs, _) in parsed.items():
        items = []
        for r in reqs:
            request = {
                "method": r["method"],
                "header": [{"key": h["key"], "value": to_postman(h["value"])} for h in r["headers"]],
                "url": to_postman(r["url"]),
            }
            if r["body"]:
                request["body"] = {"mode": "raw", "raw": to_postman(r["body"])}
                if any(h["key"].lower() == "content-type" and "json" in h["value"] for h in r["headers"]):
                    request["body"]["options"] = {"raw": {"language": "json"}}
            item = {"name": r["title"], "request": request}
            if r["name"] in refs:
                script = ["const json = pm.response.json();"] + [
                    f'pm.collectionVariables.set("{v}", json{p[1:]});'
                    for v, p in refs[r["name"]].items()
                ]
                item["event"] = [{"listen": "test", "script": {"type": "text/javascript", "exec": script}}]
            items.append(item)
        folders.append({"name": f.stem, "item": items})

    collection = {
        "info": {
            "name": Path.cwd().name,
            "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        },
        "item": folders,
        "variable": [{"key": k, "value": str(v)} for k, v in variables.items()]
        + [{"key": v, "value": ""} for vs in refs.values() for v in vs],
    }
    OUT.write_text(json.dumps(collection, indent=2) + "\n")
    print(f"wrote {OUT} ({sum(len(f['item']) for f in folders)} requests)")


if __name__ == "__main__":
    main()
