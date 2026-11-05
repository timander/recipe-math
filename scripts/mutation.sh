#!/usr/bin/env bash
# Mutation testing for the Go engine with Gremlins. Gremlins plants small
# bugs (flipped conditions, off-by-one boundaries, changed arithmetic) and
# checks that some test fails for each one. Results: reports/mutation.json
# and a markdown summary at reports/mutation.md.
set -euo pipefail
cd "$(dirname "$0")/.."
GREMLINS="${GREMLINS:-$(go env GOPATH)/bin/gremlins}"
mkdir -p reports
"$GREMLINS" unleash \
  --invert-logical --invert-loopctrl --invert-assignments --remove-self-assignments \
  --exclude-files 'cmd/' --exclude-files 'node_modules/' \
  --threshold-efficacy "${THRESHOLD:-0}" \
  --output reports/mutation.json --silent . || status=$?
node scripts/mutation-summary.js reports/mutation.json > reports/mutation.md
cat reports/mutation.md
exit "${status:-0}"
