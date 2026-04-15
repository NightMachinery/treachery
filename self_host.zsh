#!/usr/bin/env zsh

setopt ERR_EXIT PIPE_FAIL NO_UNSET

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$ROOT_DIR"

NODE_VERSION="${NODE_VERSION:-$(<"$ROOT_DIR/UI/.nvmrc")}"
DEFAULT_PUBLIC_URL="${DEFAULT_PUBLIC_URL:-https://treachery.pinky.lilf.ir}"
CONFIG_FILE="${CONFIG_FILE:-$ROOT_DIR/.self_host/config.env}"
CADDYFILE_PATH="${CADDYFILE_PATH:-$HOME/Caddyfile}"
APP_ADDR="${APP_ADDR:-127.0.0.1:18083}"
SESSION_APP="treachery-self-host"
CADDY_BLOCK_BEGIN="# BEGIN treachery self-host"
CADDY_BLOCK_END="# END treachery self-host"
CADDY_GLOBAL_BLOCK_BEGIN="# BEGIN treachery self-host global"
CADDY_GLOBAL_BLOCK_END="# END treachery self-host global"

SELF_HOST_DIR="$ROOT_DIR/.self_host"
BINARY_PATH="$SELF_HOST_DIR/bin/treachery-server"
DATA_DIR="$SELF_HOST_DIR/data"
DIST_DIR="$ROOT_DIR/UI/dist/deceptiongame"
CARDS_PATH="$ROOT_DIR/UI/src/assets/cards.json"

PUBLIC_URL=""
PUBLIC_HOST=""

tmuxnew () {
	tmux kill-session -t "$1" &> /dev/null || true
	tmux new -d -s "$@"
}

say() {
  print -r -- "$@"
}

die() {
  print -u2 -r -- "$@"
  exit 1
}

normalize_url() {
  local value="${1:-$DEFAULT_PUBLIC_URL}"
  value="${value%%\?*}"
  value="${value%%\#*}"
  if [[ "$value" != http://* && "$value" != https://* ]]; then
    value="https://$value"
  fi
  value="${value%/}"
  print -r -- "$value"
}

host_from_url() {
  local value
  value="$(normalize_url "$1")"
  value="${value#http://}"
  value="${value#https://}"
  print -r -- "${value%%/*}"
}

write_config() {
  local requested_url host
  requested_url="$(normalize_url "${1:-$DEFAULT_PUBLIC_URL}")"
  host="$(host_from_url "$requested_url")"
  [[ -n "$host" ]] || die "Could not determine host from $requested_url"

  mkdir -p "$SELF_HOST_DIR" "$SELF_HOST_DIR/bin" "$DATA_DIR"
  cat > "$CONFIG_FILE" <<CONFIG
PUBLIC_URL=$requested_url
PUBLIC_HOST=$host
APP_ADDR=$APP_ADDR
SESSION_APP=$SESSION_APP
BINARY_PATH=$BINARY_PATH
DATA_DIR=$DATA_DIR
DIST_DIR=$DIST_DIR
CARDS_PATH=$CARDS_PATH
CONFIG
  say "Wrote $CONFIG_FILE"
}

load_config() {
  [[ -f "$CONFIG_FILE" ]] || die "Missing $CONFIG_FILE. Run ./self_host.zsh setup [url] first."
  set -a
  source "$CONFIG_FILE"
  set +a
  PUBLIC_URL="${PUBLIC_URL:-$DEFAULT_PUBLIC_URL}"
  PUBLIC_HOST="${PUBLIC_HOST:-$(host_from_url "$PUBLIC_URL")}"
  APP_ADDR="${APP_ADDR:-127.0.0.1:18083}"
  BINARY_PATH="${BINARY_PATH:-$SELF_HOST_DIR/bin/treachery-server}"
  DATA_DIR="${DATA_DIR:-$SELF_HOST_DIR/data}"
  DIST_DIR="${DIST_DIR:-$ROOT_DIR/UI/dist/deceptiongame}"
  CARDS_PATH="${CARDS_PATH:-$ROOT_DIR/UI/src/assets/cards.json}"
}

ensure_tools() {
  command -v tmux >/dev/null 2>&1 || die "tmux is required"
  command -v caddy >/dev/null 2>&1 || die "caddy is required"
  command -v pnpm >/dev/null 2>&1 || die "pnpm is required"
  command -v go >/dev/null 2>&1 || die "go is required"
}

load_node() {
  nvm-load
  nvm use "$NODE_VERSION"
}

proxy_exports() {
  local names=(ALL_PROXY all_proxy http_proxy https_proxy HTTP_PROXY HTTPS_PROXY npm_config_proxy npm_config_https_proxy NO_PROXY no_proxy)
  local name value
  for name in $names; do
    if (( ${+parameters[$name]} )); then
      value="${(P)name}"
      print -r -- "export $name=${(q)value};"
    fi
  done
}

ensure_port_free() {
  local port="${APP_ADDR##*:}"
  if ss -ltn | awk '{print $4}' | grep -Eq "(^|:)$port$"; then
    die "Required listen port $port is already in use."
  fi
}

ensure_build_artifacts() {
  [[ -x "$BINARY_PATH" ]] || die "Missing $BINARY_PATH. Run ./self_host.zsh setup or redeploy first."
  [[ -f "$DIST_DIR/index.html" ]] || die "Missing $DIST_DIR/index.html. Run ./self_host.zsh setup or redeploy first."
  [[ -f "$CARDS_PATH" ]] || die "Missing $CARDS_PATH"
}

build_ui() {
  say "Installing UI dependencies with pnpm"
  (
    cd "$ROOT_DIR/UI"
    load_node
    pnpm install --frozen-lockfile --prefer-offline
    pnpm build
  )
}

build_server() {
  mkdir -p "$SELF_HOST_DIR/bin" "$DATA_DIR"
  say "Running Go tests"
  go test ./server/...
  say "Building Go server"
  go build -o "$BINARY_PATH" ./server/cmd/treachery-server
}

write_caddy_block() {
  local app_port="${APP_ADDR##*:}"
  python3 - "$CADDYFILE_PATH" "$PUBLIC_HOST" "$app_port" "$CADDY_BLOCK_BEGIN" "$CADDY_BLOCK_END" "$CADDY_GLOBAL_BLOCK_BEGIN" "$CADDY_GLOBAL_BLOCK_END" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1]).expanduser()
host = sys.argv[2]
port = sys.argv[3]
block_begin = sys.argv[4]
block_end = sys.argv[5]
global_begin = sys.argv[6]
global_end = sys.argv[7]

site_block = f"""{block_begin}
http://{host} {{
    encode zstd gzip
    reverse_proxy 127.0.0.1:{port}
}}

https://{host} {{
    tls internal
    encode zstd gzip
    reverse_proxy 127.0.0.1:{port}
}}
{block_end}
"""

global_block = f"""{global_begin}
{{
    auto_https disable_redirects
}}
{global_end}
"""

def remove_marked_block(text: str, begin: str, finish: str) -> str:
    if begin in text and finish in text:
        start_idx = text.index(begin)
        end_idx = text.index(finish, start_idx) + len(finish)
        return text[:start_idx] + text[end_idx:]
    return text

try:
    text = path.read_text()
except FileNotFoundError:
    text = ""

text = remove_marked_block(text, global_begin, global_end)
text = remove_marked_block(text, block_begin, block_end)
text = text.rstrip()

existing_global_block = any(line.strip() == "{" for line in text.splitlines())

if text:
    if existing_global_block:
        text = text + "\n\n" + site_block
    else:
        text = global_block + "\n\n" + text + "\n\n" + site_block
else:
    text = global_block + "\n\n" + site_block

if not text.endswith("\n"):
    text += "\n"

path.write_text(text)
PY
  say "Updated $CADDYFILE_PATH for $PUBLIC_HOST"
}

reload_or_start_caddy() {
  caddy validate --config "$CADDYFILE_PATH" --adapter caddyfile >/dev/null || die "Caddy validation failed for $CADDYFILE_PATH"
  if caddy reload --config "$CADDYFILE_PATH" --adapter caddyfile >/dev/null 2>&1; then
    say "Reloaded Caddy from $CADDYFILE_PATH"
    return
  fi
  if pgrep -x caddy >/dev/null 2>&1; then
    die "Caddy is running, but reload failed. Try: caddy reload --config $CADDYFILE_PATH --adapter caddyfile"
  fi
  local caddy_log="$HOME/.caddy-treachery.log"
  caddy run --config "$CADDYFILE_PATH" --adapter caddyfile >"$caddy_log" 2>&1 < /dev/null &!
  say "Started Caddy with $CADDYFILE_PATH (log: $caddy_log)"
}

start_session() {
  ensure_port_free
  ensure_build_artifacts
  local exports cmd
  exports="$(proxy_exports)"
  cmd="$exports cd ${(q)ROOT_DIR}; ${(q)BINARY_PATH} -addr ${(q)APP_ADDR} -dist-dir ${(q)DIST_DIR} -data-dir ${(q)DATA_DIR} -cards-path ${(q)CARDS_PATH}"
  tmuxnew "$SESSION_APP" zsh -lc "$cmd"
  say "Started tmux session $SESSION_APP"
  say "Primary URL: $PUBLIC_URL"
  say "Also available at: http://$PUBLIC_HOST"
}

stop_session() {
  tmux kill-session -t "$SESSION_APP" &> /dev/null || true
  say "Stopped tmux session $SESSION_APP"
}

prepare_config() {
  if [[ -n "${1:-}" ]]; then
    write_config "$1"
  elif [[ ! -f "$CONFIG_FILE" ]]; then
    write_config "$DEFAULT_PUBLIC_URL"
  fi
  load_config
}

setup_cmd() {
  stop_session
  prepare_config "${1:-}"
  ensure_tools
  build_ui
  build_server
  write_caddy_block
  reload_or_start_caddy
  start_session
}

redeploy_cmd() {
  stop_session
  prepare_config "${1:-}"
  ensure_tools
  build_ui
  build_server
  write_caddy_block
  reload_or_start_caddy
  start_session
}

start_cmd() {
  prepare_config "${1:-}"
  ensure_tools
  write_caddy_block
  reload_or_start_caddy
  start_session
}

main() {
  case "${1:-}" in
    setup)
      setup_cmd "${2:-}"
      ;;
    redeploy)
      redeploy_cmd "${2:-}"
      ;;
    start)
      start_cmd "${2:-}"
      ;;
    stop)
      stop_session
      ;;
    *)
      die "Usage: ./self_host.zsh [setup|redeploy|start|stop] [url]"
      ;;
  esac
}

main "$@"
