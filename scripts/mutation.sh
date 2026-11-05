#!/usr/bin/env bash
# Mutation testing for the Go engine with Gremlins. Gremlins plants small
# bugs (flipped conditions, off-by-one boundaries, changed arithmetic) and
# checks that some test fails for each one.
#
#   ./scripts/mutation.sh                         every file → reports/mutation.{json,md}
#   FILES="kitchen/orders.go" OUT=reports/x.json  one shard (CI runs shards in parallel)
#   THRESHOLD=90                                  fail under 90% efficacy
set -euo pipefail
cd "$(dirname "$0")/.."
GREMLINS="${GREMLINS:-$(go env GOPATH)/bin/gremlins}"
OUT="${OUT:-reports/mutation.json}"
mkdir -p "$(dirname "$OUT")"

# A shard mutates only FILES, so exclude every other engine source file.
excludes=(--exclude-files 'cmd/' --exclude-files 'node_modules/')
if [[ -n "${FILES:-}" ]]; then
  for f in frac/*.go kitchen/*.go; do
    [[ $f == *_test.go || " $FILES " == *" $f "* ]] && continue
    excludes+=(--exclude-files "${f//./\\.}\$")
  done
fi

# Gremlins gives each mutant 'timeout coefficient' × its first test run to
# finish. Some planted bugs loop forever, so that first run must measure the
# tests alone: compile everything now, and turn off Go's test-result cache
# (otherwise the baseline is near zero and every mutant "times out").
export GOFLAGS=-count=1
go test -cover ./frac ./kitchen > /dev/null

"$GREMLINS" unleash \
  --invert-logical --invert-loopctrl --invert-assignments --remove-self-assignments \
  "${excludes[@]}" \
  --timeout-coefficient "${TIMEOUT_COEFFICIENT:-10}" \
  --output "$OUT" --silent . || true
node scripts/mutation-summary.js --threshold "${THRESHOLD:-0}" "$OUT" > "${OUT%.json}.md" || status=$?
cat "${OUT%.json}.md"
exit "${status:-0}"
