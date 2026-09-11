#!/usr/bin/env bash
set -u

# Run on either load balancer while HAProxy is deliberately stopped on the
# current VIP owner in another terminal. This script does not perform failover;
# it only records observable HTTP behaviour with UTC timestamps.
url="${1:-http://10.77.0.10:8080/readyz}"
success=0
failure=0
current_failure_streak=0
max_failure_streak=0

print_final_count() {
  printf 'FINAL success=%d failure=%d max_consecutive_failures=%d\n' \
    "$success" "$failure" "$max_failure_streak"
}

trap 'print_final_count; exit 0' INT TERM

while true; do
  timestamp="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  if curl --silent --show-error --fail --max-time 2 "$url" >/dev/null; then
    success=$((success + 1))
    current_failure_streak=0
    result=200
  else
    failure=$((failure + 1))
    current_failure_streak=$((current_failure_streak + 1))
    if (( current_failure_streak > max_failure_streak )); then
      max_failure_streak=$current_failure_streak
    fi
    result=error
  fi
  printf '%s result=%s success=%d failure=%d consecutive_failures=%d\n' \
    "$timestamp" "$result" "$success" "$failure" \
    "$current_failure_streak"
  sleep 1
done
