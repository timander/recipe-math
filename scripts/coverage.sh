#!/usr/bin/env bash
# Go test coverage: reports/coverage.out, an HTML report at
# reports/coverage.html, and a markdown summary at reports/coverage.md.
# Fails if total coverage drops below MIN_COVERAGE (default 90).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p reports
go test -covermode=atomic -coverprofile=reports/coverage.out ./frac ./kitchen
go tool cover -html=reports/coverage.out -o reports/coverage.html
{
  echo "## Go test coverage"
  echo
  echo "| Package / function | Coverage |"
  echo "|---|---:|"
  go tool cover -func=reports/coverage.out | awk '
    /^total:/ { total = $3; next }
    { split($1, p, ":"); pkg = p[1]; sub(/\/[^\/]+$/, "", pkg); sub(/.*recipe-math\//, "", pkg)
      if ($3 + 0 < 80) low = low "| " pkg " " $2 " | " $3 " |\n" }
    END { print "| **total** | **" total "** |"; if (low != "") printf "%s", low }'
  echo
  echo "Functions under 80% are listed individually."
} > reports/coverage.md
cat reports/coverage.md
total=$(go tool cover -func=reports/coverage.out | awk '/^total:/ { sub("%", "", $3); print $3 }')
min="${MIN_COVERAGE:-90}"
awk -v t="$total" -v m="$min" 'BEGIN { if (t + 0 < m + 0) { printf "coverage %.1f%% is below %s%%\n", t, m; exit 1 } }'
