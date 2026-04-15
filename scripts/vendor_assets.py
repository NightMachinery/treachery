#!/usr/bin/env python3
import concurrent.futures
import hashlib
import html
import json
import mimetypes
import os
import re
import sys
from pathlib import Path
from typing import Iterable, Optional, Tuple
from urllib.parse import urlparse
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
CARDS_JSON = ROOT / 'UI' / 'src' / 'assets' / 'cards.json'
CARDS_DIR = ROOT / 'UI' / 'src' / 'assets' / 'cards'
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


def download_card_image(section: str, index: int, card: dict) -> Tuple[str, str]:
    slug = slugify(card['name'])
    out_dir = CARDS_DIR / section
    out_dir.mkdir(parents=True, exist_ok=True)
    base_name = f"{index + 1:03d}-{slug}"

    for candidate_url in [card.get('imgUrl', ''), card.get('altImgUrl', '')]:
        if not candidate_url:
            continue
        try:
            data, content_type = fetch_bytes(candidate_url)
            if not content_type.startswith('image/'):
                continue
            ext = extension_for(candidate_url, content_type)
            out_path = out_dir / f"{base_name}{ext}"
            out_path.write_bytes(data)
            rel = out_path.relative_to(ROOT / 'UI' / 'src').as_posix()
            return rel, candidate_url
        except Exception as exc:
            print(f"warn: failed {candidate_url}: {exc}", file=sys.stderr)
            continue

    out_path = out_dir / f"{base_name}.svg"
    out_path.write_bytes(placeholder_svg(card['name'], section))
    rel = out_path.relative_to(ROOT / 'UI' / 'src').as_posix()
    return rel, 'placeholder'


def vendor_cards() -> None:
    data = json.loads(CARDS_JSON.read_text())
    jobs = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
        for section in ['clueCards', 'meansCards']:
            for index, card in enumerate(data[section]):
                jobs.append((section, index, card, executor.submit(download_card_image, section.replace('Cards', ''), index, card)))

        for section, index, card, future in jobs:
            rel_path, source = future.result()
            card['imgUrl'] = rel_path
            card['altImgUrl'] = rel_path
            print(f"card {section}[{index}] -> {rel_path} ({source})")

    CARDS_JSON.write_text(json.dumps(data, indent=2) + '\n')


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
