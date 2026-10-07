#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
mkdir -p http
rm -f http/.delete

port=$(awk '/^server:/{s=1} s&&/port:/{print $2; exit}' config.yaml)

if [[ ! -f http/http-client.env.json ]]; then
  cat > http/http-client.env.json <<EOF
{
  "dev": {
    "base": "http://localhost:${port}"
  }
}
EOF
fi

[[ -d http/.git ]] || git -C http init -q
if [[ -n $(git -C http status --porcelain) ]]; then
  git -C http add -A
  git -C http commit -qm "manual edits"
fi

read -r -d '' prompt <<'EOF' || true
Sync the kulala.nvim .http files in ./http with the Go HTTP handlers in this project.

./http is its own git repo. Commits named "sync" are your previous runs; commits named
"manual edits" are changes the user made by hand. Start by running
`git -C http log --stat -10` and `git -C http log -p -3`. Treat hand edits as intentional:
keep the user's renamed requests, body values, extra requests, headers and ordering
unless the route they target no longer exists or its method/path/fields changed.

1. Find every route registered in the Go code (mux.HandleFunc, Handle, etc.) and read
   each handler to learn its method, path, path params, request body fields (json tags),
   and possible status codes.
2. Keep one .http file per Go package that registers routes, named after the resource
   (e.g. internal/host -> http/hosts.http).
3. In each file, one request per route, separated by "### <short name>".
   - Use {{base}} for the host (defined in http/http-client.env.json, do not edit it).
   - Give requests that create a resource a "# @name <name>" line, and have later
     requests reuse the id via {{<name>.response.body.$.id}} instead of hardcoded ids.
   - Request bodies use realistic example values for every json field the handler reads.
   - Order: create, list, get, update, delete.
   - No comments other than the ### separators and @name lines.
4. Update existing files in place: add requests for new routes, fix ones whose method,
   path or body fields changed, remove requests for routes that no longer exist.
5. If a whole .http file no longer matches any package, do not delete it yourself.
   Instead write its filename (one per line) to http/.delete.

Only write inside ./http. Never modify Go code or anything else. Do not commit.
End with a short plain list of what you added, changed or removed.
EOF

claude -p "$prompt" \
  --allowedTools "Read" "Glob" "Grep" "Edit(./http/**)" \
    "Bash(git -C http log:*)" "Bash(git -C http diff:*)" "Bash(git -C http show:*)" "Bash(git -C http status:*)"

if [[ -f http/.delete ]]; then
  while IFS= read -r f; do
    f=$(basename "$f")
    [[ "$f" == *.http && -f "http/$f" ]] && rm -v "http/$f"
  done < http/.delete
  rm -f http/.delete
fi

if [[ -n $(git -C http status --porcelain) ]]; then
  git -C http add -A
  git -C http commit -qm "sync"
fi
git -C http log --oneline -3
