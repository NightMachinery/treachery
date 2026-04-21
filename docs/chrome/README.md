# Remote Chrome debugging playbook

This folder documents the repeatable workflow we use when debugging Treachery with the remote Chrome MCP browser.

Use this when you want to:

- inspect the real UI in a remote browser
- debug desktop and mobile layouts
- debug multiplayer or role-specific flows
- test image-card mode and text-only means/clues mode
- create disposable rooms instead of mutating a real room by hand

## 1. Prefer `dev-start` while iterating on UI

For frontend debugging, use the development mode launcher instead of redeploying:

```bash
./self_host.zsh dev-start treachery.pinky.lilf.ir
```

Why:

- `/api/*` still goes to the Go server
- non-API requests go to Angular's dev server
- CSS/template/component changes hot-reload much faster than a redeploy

`start`, `dev-start`, and `redeploy` already stop both managed sessions before starting a new one.

## 2. Snapshot first, screenshot second

When using the Chrome MCP server:

- prefer `take_snapshot` for structure, labels, roles, and interactable elements
- use `take_screenshot` only when visual polish actually matters
- after each significant UI change, re-run both on desktop and mobile

Good default loop:

1. `take_snapshot`
2. identify layout defects
3. patch code
4. wait for Angular dev server to recompile
5. `navigate_page` or open a fresh page
6. `take_snapshot`
7. `take_screenshot`

## 3. Use isolated browser contexts for multiplayer

For Treachery, one browser tab is often not enough.

Best practice:

- one `isolatedContext` per fake player/session
- start each context from `about:blank`
- inject auth before navigation using `initScript`
- use separate pages for:
  - murderer/player view
  - forensic scientist view
  - observer view
  - mobile variants

This prevents session leakage between roles.

## 4. Seed disposable debug rooms instead of clicking through setup every time

Use the helper script in this folder:

```bash
python docs/chrome/create_debug_rooms.py \
  --public-origin http://treachery.pinky.lilf.ir \
  --connect-origin http://127.0.0.1 \
  --host-header treachery.pinky.lilf.ir
```

What it does:

- creates two disposable 4-player games
- starts both games
- creates one room in normal image-card mode
- creates one room in text-only means/clues mode
- auto-selects murderer cards
- auto-reveals a usable set of forensic clues
- prints JSON with:
  - game IDs
  - players and role hints
  - ready-to-paste `initScript` snippets
  - recommended routes for each player

This is the fastest path to role-specific UI testing.

## 5. Auth bootstrap pattern for Chrome pages

If you already have a session token, open a fresh page with an init script like this:

```js
(() => {
  document.cookie = 'treachery_session=SESSION_TOKEN; path=/; SameSite=Lax';
  localStorage.setItem('treachery.session.displayName', 'Alpha');
})();
```

Then navigate directly to the target route, for example:

- `/play/ABCD`
- `/forensic/ABCD`
- `/observe/ABCD`

The helper script prints this snippet for each seeded player.

A reusable template also lives at `docs/chrome/session_bootstrap.js`.

## 6. Desktop/mobile verification checklist

Always test both desktop and mobile.

### Desktop

Recommended checks:

- overall section hierarchy
- chat/sidebar balance
- suspect deck readability
- forensic clue density
- button spacing and panel rhythm
- image cards vs text-only cards

### Mobile

Always enable mobile emulation explicitly.

Recommended checks:

- navbar height and wrapping
- timer placement
- chat placement below content
- suspect deck stacking
- forensic clue readability without side scrolling
- button tap targets
- text-only card wrapping

## 7. When to open a fresh page instead of reusing one

Open a new isolated page when:

- a page is stuck on the wrong route
- the session became stale or anonymous again
- reloads are timing out
- websocket/live-reload state got weird
- you want a second role or device view

In practice, fresh pages are often faster than repairing a stale one.

## 8. Recommended debugging workflow for Treachery UI work

1. run `./self_host.zsh dev-start ...`
2. generate disposable rooms with `create_debug_rooms.py`
3. open a fresh isolated Chrome page per role you care about
4. inspect desktop image-card mode
5. inspect desktop text-only mode
6. enable mobile emulation and re-check both
7. patch UI
8. wait for Angular compile success
9. re-open or reload fresh pages
10. take final snapshots/screenshots
11. update docs if the workflow or UI behavior changed

## 9. Troubleshooting

### Anonymous user even though you injected a token

Use a fresh `about:blank` page and pass the session bootstrap code via `initScript` during navigation.

### Shell requests hit a proxy or return `503`

When scripting locally on the VPS, prefer `--connect-origin http://127.0.0.1` and pass the public host via `--host-header ...`.

That avoids proxy interference while still hitting the correct app.

### Dev server recompiles but Chrome still looks stale

Open a fresh page in a new isolated context instead of depending on a problematic reload.

### Live-reload console spam

Webpack/ng-serve reconnect messages are normal during `dev-start`; focus on whether the actual page content updated.

## 10. Files in this folder

- `README.md`: this playbook
- `create_debug_rooms.py`: disposable seeded-room generator for UI debugging
- `session_bootstrap.js`: reusable auth bootstrap template for Chrome MCP `initScript`
