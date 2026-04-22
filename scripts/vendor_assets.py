#!/usr/bin/env python3
import concurrent.futures
import html
import mimetypes
import os
import re
import sys
from pathlib import Path
from typing import Iterable, Optional, Tuple
from urllib.parse import urlparse
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
CRIME_PACK_DIR = ROOT / 'wordpacks' / 'crime' / 'treachery'
ASSET_SET_ID = 'treachery'
MEANS_IDS = CRIME_PACK_DIR / 'means' / 'ids.txt'
MEANS_EN = CRIME_PACK_DIR / 'means' / 'languages' / 'en.txt'
CLUES_IDS = CRIME_PACK_DIR / 'clues' / 'ids.txt'
CLUES_EN = CRIME_PACK_DIR / 'clues' / 'languages' / 'en.txt'
ASSETS_DIR = CRIME_PACK_DIR / 'assets' / ASSET_SET_ID
FONTS_DIR = ROOT / 'UI' / 'src' / 'assets' / 'fonts'
FONTS_SCSS = ROOT / 'UI' / 'src' / '_fonts.scss'
USER_AGENT = 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0 Safari/537.36'
TIMEOUT = 45
ROBOTO_CSS_URL = 'https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;700&display=swap'

EXT_BY_CONTENT_TYPE = {
    'image/jpeg': '.jpg',
    'image/jpg': '.jpg',
    'image/png': '.png',
    'image/webp': '.webp',
    'image/gif': '.gif',
    'image/svg+xml': '.svg',
    'image/avif': '.avif',
    'image/bmp': '.bmp',
    'font/woff2': '.woff2',
}


def slugify(value: str) -> str:
    value = value.lower().strip()
    value = re.sub(r'[^a-z0-9]+', '-', value)
    return value.strip('-') or 'card'


def fetch_bytes(url: str) -> Tuple[bytes, str]:
    request = Request(url, headers={'User-Agent': USER_AGENT})
    with urlopen(request, timeout=TIMEOUT) as response:
        content_type = response.headers.get_content_type()
        return response.read(), content_type


def extension_for(url: str, content_type: str) -> str:
    if content_type in EXT_BY_CONTENT_TYPE:
        return EXT_BY_CONTENT_TYPE[content_type]
    parsed = urlparse(url)
    suffix = Path(parsed.path).suffix.lower()
    if suffix:
        return suffix
    guessed = mimetypes.guess_extension(content_type)
    return guessed or '.bin'


def placeholder_svg(card_name: str, section: str) -> bytes:
    escaped = html.escape(card_name)
    subtitle = html.escape(section.title())
    svg = f"""<svg xmlns='http://www.w3.org/2000/svg' width='640' height='900' viewBox='0 0 640 900'>
  <rect width='640' height='900' fill='#202020'/>
  <rect x='28' y='28' width='584' height='844' rx='28' fill='#2f2f2f' stroke='#7e56c1' stroke-width='8'/>
  <text x='320' y='320' text-anchor='middle' fill='#ffffff' font-size='44' font-family='Roboto, Helvetica, Arial, sans-serif'>{escaped}</text>
  <text x='320' y='410' text-anchor='middle' fill='#d8ccec' font-size='28' font-family='Roboto, Helvetica, Arial, sans-serif'>{subtitle}</text>
  <text x='320' y='520' text-anchor='middle' fill='#aaaaaa' font-size='24' font-family='Roboto, Helvetica, Arial, sans-serif'>Local placeholder image</text>
</svg>"""
    return svg.encode('utf-8')

def load_lines(path: Path) -> list[str]:
    return [line.strip() for line in path.read_text(encoding='utf-8').splitlines() if line.strip()]


def iter_pack_cards(deck: str) -> list[tuple[str, str]]:
    if deck == 'means':
        ids = load_lines(MEANS_IDS)
        labels = load_lines(MEANS_EN)
    else:
        ids = load_lines(CLUES_IDS)
        labels = load_lines(CLUES_EN)
    if len(ids) != len(labels):
        raise RuntimeError(f'{deck}: ids/labels length mismatch')
    return list(zip(ids, labels))


def ensure_pack_image(deck: str, card_id: str, label: str) -> tuple[str, str]:
    out_dir = ASSETS_DIR / deck
    out_dir.mkdir(parents=True, exist_ok=True)
    matches = sorted(out_dir.glob(f'{card_id}.*'))
    if matches:
        return matches[0].relative_to(ROOT).as_posix(), 'existing'
    out_path = out_dir / f'{card_id}.svg'
    out_path.write_bytes(placeholder_svg(label, deck))
    return out_path.relative_to(ROOT).as_posix(), 'placeholder'


def vendor_cards() -> None:
    jobs = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
        for deck in ['clues', 'means']:
            for card_id, label in iter_pack_cards(deck):
                jobs.append((deck, card_id, label, executor.submit(ensure_pack_image, deck, card_id, label)))

        for deck, card_id, _label, future in jobs:
            rel_path, source = future.result()
            print(f'card {deck}/{card_id} -> {rel_path} ({source})')


def fetch_text(url: str, accept: Optional[str] = None) -> str:
    headers = {'User-Agent': USER_AGENT}
    if accept:
        headers['Accept'] = accept
    request = Request(url, headers=headers)
    with urlopen(request, timeout=TIMEOUT) as response:
        return response.read().decode('utf-8')


def vendor_fonts() -> None:
    FONTS_DIR.mkdir(parents=True, exist_ok=True)
    css = fetch_text(ROBOTO_CSS_URL, 'text/css,*/*;q=0.1')
    weight_to_url = {}
    current_weight = None
    for line in css.splitlines():
        line = line.strip()
        if line.startswith('font-weight:'):
            current_weight = line.split(':', 1)[1].strip().rstrip(';')
        elif 'src:' in line and 'url(' in line and current_weight:
            match = re.search(r'url\(([^)]+)\)', line)
            if match:
                weight_to_url[current_weight] = match.group(1)
                current_weight = None

    if not weight_to_url:
        raise RuntimeError('Could not parse Roboto font URLs from Google Fonts CSS')

    lines = []
    for weight, url in sorted(weight_to_url.items(), key=lambda item: int(item[0])):
        data, content_type = fetch_bytes(url)
        ext = extension_for(url, content_type)
        file_name = f'roboto-{weight}{ext}'
        (FONTS_DIR / file_name).write_bytes(data)
        lines.extend([
            '@font-face {',
            "  font-family: 'Roboto';",
            "  font-style: normal;",
            f'  font-weight: {weight};',
            '  font-display: swap;',
            f"  src: url('/assets/fonts/{file_name}') format('woff2');",
            '}',
            '',
        ])
    lines.extend([
        'body {',
        "  font-family: 'Roboto', 'Helvetica Neue', Arial, sans-serif;",
        '}',
        '',
    ])
    FONTS_SCSS.write_text('\n'.join(lines))


def main() -> int:
    vendor_cards()
    vendor_fonts()
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
