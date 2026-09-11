#!/usr/bin/env bash
set -u

# This script is opened and explained before it is executed. It does not hide
# the upgrade: it only records whether the already published /readyz endpoint
# answered while the official Kubespray playbook was running.
url="${1:-http://127.0.0.1:18080/readyz}"
success=0
failure=0
current_failure_streak=0
max_failure_streak=0

print_final_count() {
  printf 'FINAL success=%d failure=%d max_consecutive_failures=%d\n' \
    "$success" "$failure" "$max_failure_streak"
}

# Ctrl-C не теряет итог: перед завершением скрипт печатает оба счётчика.
trap 'print_final_count; exit 0' INT TERM

while true; do
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
    "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$result" "$success" "$failure" \
    "$current_failure_streak"
  sleep 1
done
