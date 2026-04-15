# Treachery self-hosting

This repo now self-hosts without Firebase or Docker.

## What runs

- **Caddy** terminates HTTP / self-signed HTTPS for `treachery.pinky.lilf.ir`.
- **One Go server** runs in tmux and serves:
  - the built Angular SPA
  - local anonymous session auth
  - REST APIs
  - live game updates over SSE
  - SQLite-backed game state in `.self_host/data/treachery.sqlite`

Default origin:

- primary: `https://treachery.pinky.lilf.ir`
- also available on plain HTTP: `http://treachery.pinky.lilf.ir`

## Intranet/offline notes

- All runtime card images are committed under `UI/src/assets/cards/`.
- Roboto fonts are committed under `UI/src/assets/fonts/`.
- The viewport fix script is local at `UI/src/assets/vh-fix.js`.
- No Google Fonts, Firebase, reCAPTCHA, or other required external runtime services remain.
- Clipboard copy uses a `navigator.clipboard` path when available and an `execCommand('copy')` fallback so HTTP still works.

## Commands

```zsh
./self_host.zsh setup [url]
./self_host.zsh redeploy [url]
./self_host.zsh start [url]
./self_host.zsh stop
```

`[url]` may be a host like `treachery.pinky.lilf.ir` or a full origin like `https://treachery.pinky.lilf.ir`.
If omitted, the default is `https://treachery.pinky.lilf.ir`.

## Command behavior

### `setup`

1. Stops the tmux session if it exists.
2. Writes `.self_host/config.env`.
3. Loads Node via:
   ```zsh
   nvm-load
   nvm use 16.20.2
   ```
4. Runs `pnpm install --frozen-lockfile --prefer-offline` in `UI/`.
5. Builds the Angular app.
6. Runs `go test ./server/...` and builds the Go server binary.
7. Updates the managed block in `~/Caddyfile`.
8. Validates and reloads Caddy.
9. Starts the Go server in tmux.

### `redeploy`

Same as `setup`, but meant for redeploying the latest local code changes.

### `start`

- Reuses the existing build artifacts and config.
- Rewrites / reloads the managed Caddy block.
- Fails fast if the app port is already taken by another process.

### `stop`

- Stops only the tmux app session.
- Leaves the managed Caddy block in place.

## Managed files and paths

- config: `.self_host/config.env`
- binary: `.self_host/bin/treachery-server`
- data: `.self_host/data/treachery.sqlite`
- tmux session: `treachery-self-host`
- Caddy block markers:
  - `# BEGIN treachery self-host global`
  - `# END treachery self-host global`
  - `# BEGIN treachery self-host`
  - `# END treachery self-host`

## Proxy handling

`self_host.zsh` does **not** hardcode proxy settings.
If proxy variables such as `ALL_PROXY`, `HTTP_PROXY`, `HTTPS_PROXY`, `npm_config_proxy`, or `npm_config_https_proxy` are present in the environment, the script uses them for builds and passes them through to the tmux session.

## Verification checklist

After `setup` or `redeploy`:

```bash
curl -I http://treachery.pinky.lilf.ir
curl -k -I https://treachery.pinky.lilf.ir
ss -ltn | grep 18083
TMUX= tmux ls | grep treachery-self-host
```

Then verify in a browser:

- home page loads on HTTP and HTTPS
- creating a game works
- joining from another browser/device works
- chat live-updates work
- forensic / murderer private views stay isolated per browser session
- copy-link works on **HTTP**

## Troubleshooting

- If `18083` is already in use, stop the conflicting process before `start` / `setup`.
- If Caddy validation fails, inspect `~/Caddyfile` and retry.
- If tmux is running but the site is blank, inspect the session:
  ```bash
  tmux attach -t treachery-self-host
  ```
- If you changed frontend dependencies, rerun `./self_host.zsh redeploy` instead of `start`.
