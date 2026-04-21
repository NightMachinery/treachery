#!/usr/bin/env python3
"""Create disposable started games for remote Chrome UI debugging.

The script creates two 4-player games:
- one in normal image-card mode
- one in text-only means/clues mode

It also auto-selects murderer cards and enough forensic clues to make the
in-game UI useful for screenshots and layout debugging.
"""

from __future__ import annotations

import argparse
import http.cookiejar
import json
import random
import string
import sys
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any

DEFAULT_PLAYER_NAMES = ["Alpha", "Bravo", "Charlie", "Delta"]


@dataclass
class SessionClient:
    opener: urllib.request.OpenerDirector
    uid: str
    token: str
    name: str


class ApiClient:
    def __init__(self, connect_origin: str, public_origin: str, host_header: str | None):
        self.connect_origin = connect_origin.rstrip("/")
        self.public_origin = public_origin.rstrip("/")
        self.host_header = host_header or urllib.parse.urlparse(public_origin).netloc

    def _build_opener(self) -> urllib.request.OpenerDirector:
        jar = http.cookiejar.CookieJar()
        opener = urllib.request.build_opener(
            urllib.request.ProxyHandler({}),
            urllib.request.HTTPCookieProcessor(jar),
        )
        opener.addheaders = [("Accept", "application/json")]
        return opener

    def _request(
        self,
        opener: urllib.request.OpenerDirector,
        method: str,
        path: str,
        data: dict[str, Any] | None = None,
    ) -> tuple[int, Any]:
        headers: dict[str, str] = {}
        if self.host_header:
            headers["Host"] = self.host_header
        body = None
        if data is not None:
            body = json.dumps(data).encode("utf-8")
            headers["Content-Type"] = "application/json"

        req = urllib.request.Request(
            self.connect_origin + path,
            method=method,
            data=body,
            headers=headers,
        )
        with opener.open(req, timeout=20) as response:
            raw = response.read().decode("utf-8")
            return response.getcode(), (json.loads(raw) if raw else None)

    def create_named_session(self, display_name: str) -> SessionClient:
        opener = self._build_opener()
        _, session = self._request(opener, "POST", "/api/session", {})
        self._request(opener, "PUT", "/api/me", {"displayName": display_name})
        return SessionClient(opener=opener, uid=session["uid"], token=session["token"], name=display_name)

    def game_exists(self, opener: urllib.request.OpenerDirector, game_id: str) -> bool:
        try:
            self._request(opener, "GET", f"/api/games/{game_id}/snapshot")
            return True
        except urllib.error.HTTPError as exc:
            if exc.code == 404:
                return False
            raise

    def find_unused_game_id(self, opener: urllib.request.OpenerDirector) -> str:
        while True:
            game_id = "".join(random.choice(string.ascii_uppercase) for _ in range(4))
            if not self.game_exists(opener, game_id):
                return game_id

    def request(self, client: SessionClient, method: str, path: str, data: dict[str, Any] | None = None) -> Any:
        _, payload = self._request(client.opener, method, path, data)
        return payload


def load_card_resources() -> dict[str, Any]:
    repo_root = Path(__file__).resolve().parents[2]
    cards_path = repo_root / "UI" / "src" / "assets" / "cards.json"
    return json.loads(cards_path.read_text(encoding="utf-8"))


def build_init_script(token: str, display_name: str) -> str:
    return (
        "(() => { "
        f"document.cookie = 'treachery_session={token}; path=/; SameSite=Lax'; "
        f"localStorage.setItem('treachery.session.displayName', {json.dumps(display_name)}); "
        "})();"
    )


def suggested_route(game_id: str, viewer: dict[str, Any]) -> str:
    if viewer.get("isScientist"):
        return f"/forensic/{game_id}"
    if viewer.get("role") == "observer":
        return f"/observe/{game_id}"
    return f"/play/{game_id}"


def select_murderer_cards(api: ApiClient, murderer: SessionClient, game_id: str) -> None:
    snapshot = api.request(murderer, "GET", f"/api/games/{game_id}/snapshot")
    my_player = next(player for player in snapshot["players"] if player["uid"] == murderer.uid)
    api.request(
        murderer,
        "POST",
        f"/api/games/{game_id}/murderer-selection",
        {
            "clueCardName": my_player["clueCards"][0]["name"],
            "meansCardName": my_player["meansCards"][0]["name"],
        },
    )


def reveal_forensic_cards(api: ApiClient, scientist: SessionClient, game_id: str, cards: dict[str, Any], selected_other_count: int) -> None:
    cause_card = dict(cards["forensicCards"]["causeCards"][0])
    cause_card["selectedChoice"] = cause_card["choices"][0]

    location_card = dict(cards["forensicCards"]["locationCards"][0])
    location_card["selectedChoice"] = location_card["choices"][0]

    api.request(scientist, "POST", f"/api/games/{game_id}/forensic/cause", {"card": cause_card})
    api.request(scientist, "POST", f"/api/games/{game_id}/forensic/location", {"card": location_card})

    snapshot = api.request(scientist, "GET", f"/api/games/{game_id}/snapshot")
    chosen = 0
    for card in snapshot["game"]["otherCards"]:
        if chosen >= selected_other_count:
            break
        payload = {
            key: card[key]
            for key in ("cardName", "choices", "selectedChoice", "replaced")
            if key in card
        }
        if not payload.get("selectedChoice"):
            payload["selectedChoice"] = payload["choices"][0]
        api.request(scientist, "POST", f"/api/games/{game_id}/forensic/other", {"card": payload, "replaceCardName": ""})
        chosen += 1


def create_seeded_game(
    api: ApiClient,
    cards: dict[str, Any],
    game_label: str,
    player_names: list[str],
    text_only: bool,
    selected_other_count: int,
) -> dict[str, Any]:
    sessions = [api.create_named_session(name) for name in player_names]
    creator = sessions[0]
    game_id = api.find_unused_game_id(creator.opener)

    api.request(creator, "POST", "/api/games", {"gameId": game_id})
    for client in sessions[1:]:
        api.request(client, "POST", f"/api/games/{game_id}/join", {"role": "player"})

    api.request(creator, "POST", f"/api/games/{game_id}/start", {})
    if text_only:
        api.request(creator, "POST", f"/api/games/{game_id}/room-mods", {"meansCluesTextOnly": True})

    snapshots: list[tuple[SessionClient, dict[str, Any]]] = []
    for client in sessions:
        snapshots.append((client, api.request(client, "GET", f"/api/games/{game_id}/snapshot")))

    murderer = next(client for client, snap in snapshots if snap["playerPrivateData"].get("isMurderer"))
    scientist = next(client for client, snap in snapshots if snap["viewer"].get("isScientist"))

    select_murderer_cards(api, murderer, game_id)
    reveal_forensic_cards(api, scientist, game_id, cards, selected_other_count)

    result_players: list[dict[str, Any]] = []
    for client in sessions:
        snapshot = api.request(client, "GET", f"/api/games/{game_id}/snapshot")
        viewer = snapshot["viewer"]
        private_data = snapshot["playerPrivateData"]
        route = suggested_route(game_id, viewer)
        result_players.append(
            {
                "name": client.name,
                "uid": client.uid,
                "token": client.token,
                "route": route,
                "url": f"{api.public_origin}{route}",
                "roleHint": private_data.get("role") or ("scientist" if viewer.get("isScientist") else "investigator"),
                "isScientist": bool(viewer.get("isScientist")),
                "isMurderer": bool(private_data.get("isMurderer")),
                "chromeInitScript": build_init_script(client.token, client.name),
            }
        )

    return {
        "label": game_label,
        "gameId": game_id,
        "mode": "text-only" if text_only else "image",
        "joinUrl": f"{api.public_origin}/join/{game_id}",
        "players": result_players,
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--public-origin",
        default="http://treachery.pinky.lilf.ir",
        help="Public browser origin to emit in URLs (default: %(default)s)",
    )
    parser.add_argument(
        "--connect-origin",
        default="http://127.0.0.1",
        help="HTTP origin the script should call directly (default: %(default)s)",
    )
    parser.add_argument(
        "--host-header",
        default=None,
        help="Optional Host header override; useful when connect-origin is localhost",
    )
    parser.add_argument(
        "--players",
        nargs="+",
        default=DEFAULT_PLAYER_NAMES,
        help="Display names for the seeded players (default: %(default)s)",
    )
    parser.add_argument(
        "--selected-other-count",
        type=int,
        default=4,
        help="How many extra forensic cards to pre-select (default: %(default)s)",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    if len(args.players) < 4:
        print("Need at least 4 player names.", file=sys.stderr)
        return 2

    api = ApiClient(
        connect_origin=args.connect_origin,
        public_origin=args.public_origin,
        host_header=args.host_header,
    )
    cards = load_card_resources()

    result = {
        "publicOrigin": api.public_origin,
        "connectOrigin": api.connect_origin,
        "hostHeader": api.host_header,
        "games": {
            "image": create_seeded_game(api, cards, "Desktop/image-card debug room", args.players, False, args.selected_other_count),
            "textOnly": create_seeded_game(api, cards, "Text-only means/clues debug room", args.players, True, args.selected_other_count),
        },
    }

    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
