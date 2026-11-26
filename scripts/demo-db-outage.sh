#!/usr/bin/env bash
# Database outage demo.
#
# Proves that a PostgreSQL outage loses no events and creates no duplicates:
#   1. count rows in events
#   2. stop the postgres container
#   3. send 20 events to the API (it still answers 202, because it only needs Kafka)
#   4. wait 10 seconds, then start postgres again
#   5. wait for the consumer to catch up
#   6. check that exactly 20 new rows arrived
#
# Needs a running stack: docker compose up --build -d
# Usage: ./scripts/demo-db-outage.sh        (from the repo root, Linux, macOS or Git Bash)
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
N="${EVENTS:-20}"
CATCH_UP_TIMEOUT="${CATCH_UP_TIMEOUT:-120}"

cd "$(dirname "$0")/.."

sql() {
  docker compose exec -T postgres psql -U events_user -d events_db -tAc "$1" | tr -d '[:space:]'
}

new_uuid() {
  local h
  h=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
  printf '%s-%s-4%s-%x%s-%s\n' "${h:0:8}" "${h:8:4}" "${h:13:3}" \
    $(( (0x${h:16:1} & 0x3) | 0x8 )) "${h:17:3}" "${h:20:12}"
}

step() { printf '\n==> %s\n' "$*"; }

step "1. Counting rows in events"
before=$(sql "SELECT COUNT(*) FROM events")
echo "rows before: $before"

step "2. Stopping postgres"
pg_container=$(docker compose ps -q postgres)
docker stop "$pg_container" > /dev/null
echo "postgres stopped"

step "3. Sending $N events while the database is down"
ids=()
accepted=0
for i in $(seq 1 "$N"); do
  id=$(new_uuid)
  ids+=("'$id'")
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API_URL/v1/events" \
    -H 'Content-Type: application/json' \
    -d "{\"event_id\":\"$id\",\"event_type\":\"outage_demo\",\"payload\":{\"n\":$i}}")
  [ "$code" = "202" ] && accepted=$((accepted + 1))
  printf '  event %2d -> HTTP %s\n' "$i" "$code"
done
echo "API accepted $accepted of $N events with 202 while postgres was down"
echo "readiness during outage: HTTP $(curl -s -o /dev/null -w '%{http_code}' "$API_URL/readyz") (503 expected)"

step "4. Waiting 10 seconds (the consumer is retrying, then pausing)"
sleep 10

step "5. Starting postgres again"
docker start "$pg_container" > /dev/null
until docker compose exec -T postgres pg_isready -U events_user -d events_db > /dev/null 2>&1; do sleep 1; done
echo "postgres is accepting connections"

step "6. Waiting for the consumer to catch up (max ${CATCH_UP_TIMEOUT}s)"
id_list=$(IFS=,; echo "${ids[*]}")
arrived=0
for ((t = 0; t < CATCH_UP_TIMEOUT; t++)); do
  arrived=$(sql "SELECT COUNT(*) FROM events WHERE event_id IN ($id_list)" 2>/dev/null || echo 0)
  [ "$arrived" = "$N" ] && break
  sleep 1
done
after=$(sql "SELECT COUNT(*) FROM events")
dlq=$(sql "SELECT COUNT(*) FROM dead_letters WHERE original_key IN ($id_list)")
new_rows=$((after - before))

step "Result"
echo "rows before:            $before"
echo "rows after:             $after"
echo "new rows:               $new_rows"
echo "demo events stored:     $arrived of $N"
echo "demo events in the DLQ: $dlq"

if [ "$accepted" = "$N" ] && [ "$arrived" = "$N" ] && [ "$new_rows" = "$N" ] && [ "$dlq" = "0" ]; then
  echo "PASS: exactly $N new rows arrived (no loss, no duplicates)"
else
  echo "FAIL: expected exactly $N new rows and none in the DLQ"
  exit 1
fi
