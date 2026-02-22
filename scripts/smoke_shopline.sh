#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Smoke-test Shopline CLI command wiring without hardcoded identifiers.

Usage:
  scripts/smoke_shopline.sh [--store <name>] [--spl <path>] [--allow-mutations]

Options:
  --store <name>        Store profile name for Open API commands.
  --spl <path>          CLI executable path (default: ./spl, then PATH spl).
  --allow-mutations     Run a tiny set of live mutating checks (default: off).
  -h, --help            Show this help text.

Notes:
  - By default this script runs read checks + dry-run write checks only.
  - No store/order/product/conversation identifiers are hardcoded.
  - If .env exists in the repo root, it is loaded automatically.
EOF
}

STORE="${SHOPLINE_STORE:-}"
ALLOW_MUTATIONS=0
SPL_BIN=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --store)
      STORE="${2:-}"
      shift 2
      ;;
    --spl)
      SPL_BIN="${2:-}"
      shift 2
      ;;
    --allow-mutations)
      ALLOW_MUTATIONS=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -f ".env" ]]; then
  # shellcheck disable=SC1091
  set -a; source .env; set +a
fi

if [[ -z "$SPL_BIN" ]]; then
  if [[ -x "./spl" ]]; then
    SPL_BIN="./spl"
  elif command -v spl >/dev/null 2>&1; then
    SPL_BIN="spl"
  else
    echo "Could not find CLI binary. Build it or pass --spl <path>." >&2
    exit 2
  fi
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for smoke checks." >&2
  exit 2
fi

strip_noise() {
  sed '/^Warning: /d'
}

select_store_if_needed() {
  if [[ -n "$STORE" ]]; then
    return 0
  fi

  local auth_out names
  auth_out="$("$SPL_BIN" auth list 2>&1 | strip_noise || true)"
  mapfile -t names < <(echo "$auth_out" | awk 'NR>1 && NF>0 {print $1}')

  if [[ ${#names[@]} -eq 1 ]]; then
    STORE="${names[0]}"
    return 0
  fi

  echo "Store profile is ambiguous. Pass --store <name> (or set SHOPLINE_STORE)." >&2
  exit 2
}

STATUS_PASS=0
STATUS_WARN=0
STATUS_BLOCKED=0
STATUS_FAIL=0

mark_result() {
  local level="$1"
  local name="$2"
  local reason="${3:-}"
  case "$level" in
    PASS) ((STATUS_PASS+=1)) ;;
    WARN) ((STATUS_WARN+=1)) ;;
    BLOCKED) ((STATUS_BLOCKED+=1)) ;;
    FAIL) ((STATUS_FAIL+=1)) ;;
  esac
  if [[ -n "$reason" ]]; then
    echo "$level|$name|$reason"
  else
    echo "$level|$name"
  fi
}

classify_error() {
  local output="$1"
  if grep -qi "multiple profiles configured" <<<"$output"; then
    echo "BLOCKED:store_selection_required"
    return
  fi
  if grep -qi "Failed to obtain session cookie after 2FA" <<<"$output"; then
    echo "BLOCKED:admin_session_2fa_refresh_required"
    return
  fi
  if grep -qi "failed to open credential store" <<<"$output"; then
    echo "BLOCKED:credentials_unavailable"
    return
  fi
  if grep -qi "Failed to get tracking number after" <<<"$output"; then
    echo "WARN:no_tracking_data_for_selected_order"
    return
  fi
  if grep -qi "status code validation" <<<"$output"; then
    echo "WARN:order_state_not_eligible_for_endpoint"
    return
  fi
  echo "FAIL:unclassified_command_error"
}

run_check() {
  local name="$1"
  shift
  local output rc tag level reason
  output="$("$@" 2>&1)" && rc=0 || rc=$?
  if [[ $rc -eq 0 ]]; then
    mark_result "PASS" "$name"
    return
  fi

  tag="$(classify_error "$output")"
  level="${tag%%:*}"
  reason="${tag#*:}"
  mark_result "$level" "$name" "$reason"
}

get_json() {
  local output rc
  output="$("$@" 2>&1)" && rc=0 || rc=$?
  printf '%s\n' "$output" | strip_noise
  return "$rc"
}

select_store_if_needed

orders_json="$(get_json "$SPL_BIN" --store "$STORE" orders list --limit 1 --json --items-only)" || {
  echo "CLI command failed when fetching orders: $orders_json" >&2
  exit 1
}
order_id="$(echo "$orders_json" | jq -r '.[0].id // empty')"
order_number="$(echo "$orders_json" | jq -r '.[0].order_number // empty')"
if [[ -z "$order_id" || -z "$order_number" ]]; then
  echo "Could not resolve an order ID/number from orders list." >&2
  exit 1
fi

products_json="$(get_json "$SPL_BIN" --store "$STORE" products list --limit 1 --json --items-only)" || {
  echo "CLI command failed when fetching products: $products_json" >&2
  exit 1
}
product_id="$(echo "$products_json" | jq -r '.[0].id // empty')"
if [[ -z "$product_id" ]]; then
  echo "Could not resolve a product ID from products list." >&2
  exit 1
fi

# Read checks (live, non-mutating)
run_check "orders list" "$SPL_BIN" --store "$STORE" orders list --limit 1 --json --items-only
run_check "orders get" "$SPL_BIN" --store "$STORE" orders get "$order_id" --json
run_check "products list" "$SPL_BIN" --store "$STORE" products list --limit 1 --json --items-only
run_check "products get" "$SPL_BIN" --store "$STORE" products get "$product_id" --json
run_check "categories list" "$SPL_BIN" --store "$STORE" categories list --limit 1 --json --items-only
run_check "shipping status" "$SPL_BIN" shipping status "$order_id" --json
run_check "shipping tracking" "$SPL_BIN" shipping tracking "$order_id" --json
run_check "livestreams list" "$SPL_BIN" livestreams list --page 1 --page-size 1 --json
run_check "message-center list" "$SPL_BIN" message-center list --platform shop_messages --page 1 --page-size 1 --json

# Write checks (dry-run only)
run_check "orders cancel dry-run" "$SPL_BIN" --store "$STORE" orders cancel "$order_id" --dry-run
run_check "orders comment dry-run" "$SPL_BIN" orders comment "$order_id" --text "smoke check" --dry-run
run_check "orders admin-refund dry-run" "$SPL_BIN" orders admin-refund "$order_id" --performer-id usr_smoke --amount 1 --payment-updated-at 2026-01-01T00:00:00Z --remark smoke --dry-run
run_check "orders receipt-reissue dry-run" "$SPL_BIN" orders receipt-reissue "$order_id" --dry-run
run_check "products update dry-run" "$SPL_BIN" --store "$STORE" products update "$product_id" --title "SMOKE-CHECK" --dry-run
run_check "products update-quantity dry-run" "$SPL_BIN" --store "$STORE" products update-quantity "$product_id" --quantity 1 --dry-run
run_check "products hide dry-run" "$SPL_BIN" products hide "$product_id" --dry-run
run_check "products publish dry-run" "$SPL_BIN" products publish "$product_id" --dry-run
run_check "products unpublish dry-run" "$SPL_BIN" products unpublish "$product_id" --dry-run
run_check "shipping execute dry-run" "$SPL_BIN" shipping execute "$order_id" --order-number "$order_number" --performer-id usr_smoke --dry-run
run_check "livestreams create dry-run" "$SPL_BIN" livestreams create --title "smoke" --owner "smoke" --platform FACEBOOK --dry-run
run_check "livestreams update dry-run" "$SPL_BIN" livestreams update stream_smoke --dry-run
run_check "livestreams delete dry-run" "$SPL_BIN" livestreams delete stream_smoke --dry-run
run_check "livestreams add-products dry-run" "$SPL_BIN" livestreams add-products stream_smoke --dry-run
run_check "livestreams remove-products dry-run" "$SPL_BIN" livestreams remove-products stream_smoke --product-ids prod1 --dry-run
run_check "livestreams start dry-run" "$SPL_BIN" livestreams start stream_smoke --platform FACEBOOK --dry-run
run_check "livestreams end dry-run" "$SPL_BIN" livestreams end stream_smoke --dry-run
run_check "message-center send dry-run" "$SPL_BIN" message-center send conv_smoke --platform shop_messages --content "smoke check" --dry-run

if [[ $ALLOW_MUTATIONS -eq 1 ]]; then
  run_check "shipping print-label live" "$SPL_BIN" shipping print-label "$order_id" --json
fi

echo "SUMMARY|pass=$STATUS_PASS|warn=$STATUS_WARN|blocked=$STATUS_BLOCKED|fail=$STATUS_FAIL"
if [[ $STATUS_FAIL -gt 0 ]]; then
  exit 1
fi

