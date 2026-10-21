#!/bin/sh
# Creates the made-up project used in docs/demo.tape: sh docs/demo-project.sh DIR
set -eu
root="$1/acme-api"
rm -rf "$root"
for f in \
  cmd/server/main.go \
  internal/handlers/users.go internal/handlers/users_test.go \
  internal/store/postgres.go \
  web/package.json web/src/App.tsx web/src/main.tsx web/node_modules/react/index.js \
  .git/HEAD .github/workflows/ci.yml .env .DS_Store \
  .gitignore Dockerfile go.mod README.md; do
  mkdir -p "$root/$(dirname "$f")"
  : > "$root/$f"
done
