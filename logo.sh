#!/usr/bin/env bash
# Regenerate assets/dex-logo.svg via ../3dlogo with a FLAT wordmark on a
# tilted 3D slab.
#
# 3dlogo applies the same tilt matrix to slab and wordmark (see
# ../3dlogo/render.go:41), so the wordmark looks 3D when pitch/yaw are
# non-zero. As a workaround we run 3dlogo twice and stitch:
#   - tilted run  → keep the projected slab (and gradient defs)
#   - flat run    → keep the un-projected wordmark
# then re-center the flat wordmark on the tilted slab's bounding box.
#
# Requires:
#   - ../3dlogo built (see github.com/StephanSchmidt/3dlogo)
#   - BricolageGrotesque96ptCondensed-ExtraBold.ttf in ~/Fonts/
#   - python3

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOGO_BIN="${LOGO_BIN:-$REPO_DIR/../3dlogo/3dlogo}"
FONT="${FONT:-$HOME/Fonts/BricolageGrotesque96ptCondensed-ExtraBold.ttf}"
OUT="${OUT:-$REPO_DIR/assets/dex-logo.svg}"
GRADIENT="linear-gradient(180deg, #2d1004 0%, #c1272d 50%, #ff8c5a 100%)"

mkdir -p "$(dirname "$OUT")"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

"$LOGO_BIN" -font "$FONT" -text dex -gradient "$GRADIENT" -fg "#ffffff" \
  -out "$TMP/tilted.svg"
"$LOGO_BIN" -font "$FONT" -text dex -gradient "$GRADIENT" -fg "#ffffff" \
  -pitch 0 -yaw 0 -out "$TMP/flat.svg"

python3 - "$TMP/tilted.svg" "$TMP/flat.svg" "$OUT" <<'PY'
import re, sys

tilted_path, flat_path, out_path = sys.argv[1:4]
tilted = open(tilted_path).read()
flat = open(flat_path).read()

path_re = re.compile(r'<path d="([^"]*)" fill="([^"]*)"')

def parse_bbox(d):
    toks = re.findall(r'[A-Za-z]|-?\d+\.?\d*', d)
    xs, ys = [], []
    j = 0; cmd = ''; cx = cy = 0.0
    while j < len(toks):
        t = toks[j]
        if t.isalpha():
            cmd = t; j += 1; continue
        try:
            if cmd in 'ML':
                x = float(toks[j]); y = float(toks[j+1]); j += 2
                xs.append(x); ys.append(y); cx, cy = x, y
            elif cmd in 'ml':
                dx = float(toks[j]); dy = float(toks[j+1]); j += 2
                cx += dx; cy += dy; xs.append(cx); ys.append(cy)
            elif cmd == 'Q':
                x = float(toks[j+2]); y = float(toks[j+3]); j += 4
                xs.append(x); ys.append(y); cx, cy = x, y
            elif cmd == 'V':
                cy = float(toks[j]); j += 1; ys.append(cy); xs.append(cx)
            elif cmd == 'H':
                cx = float(toks[j]); j += 1; xs.append(cx); ys.append(cy)
            else:
                j += 1
        except (IndexError, ValueError):
            j += 1
    return min(xs), min(ys), max(xs), max(ys)

tilted_paths = path_re.findall(tilted)
flat_paths = path_re.findall(flat)
slab_d = tilted_paths[0][0]
flat_word_d, flat_word_fill = flat_paths[1]

sx0, sy0, sx1, sy1 = parse_bbox(slab_d)
fx0, fy0, fx1, fy1 = parse_bbox(flat_word_d)

slab_cx, slab_cy = (sx0 + sx1) / 2, (sy0 + sy1) / 2
flat_cx, flat_cy = (fx0 + fx1) / 2, (fy0 + fy1) / 2

# Scale the flat wordmark so its bbox height matches ~62% of the slab's
# visible height — close to the loupe wordmark/slab ratio (~113/154 ≈ 73%
# but the slab includes the depth band, so a tighter 0.62 reads better on
# a 3-letter word).
target_h = (sy1 - sy0) * 0.62
flat_h = fy1 - fy0
scale = target_h / flat_h

def transform_path(d, tx, ty, s, src_cx, src_cy):
    toks = re.findall(r'[A-Za-z]|-?\d+\.?\d*', d)
    out = []
    j = 0; cmd = ''
    while j < len(toks):
        t = toks[j]
        if t.isalpha():
            cmd = t; out.append(t); j += 1; continue
        if cmd in 'ML' and j + 1 < len(toks):
            x = float(toks[j]); y = float(toks[j+1])
            nx = tx + (x - src_cx) * s
            ny = ty + (y - src_cy) * s
            out.append(f"{nx:.4f}"); out.append(f"{ny:.4f}")
            j += 2
        else:
            out.append(toks[j]); j += 1
    s_out = ''
    last_was_num = False
    for tok in out:
        if tok.isalpha():
            s_out += tok; last_was_num = False
        else:
            if last_was_num:
                s_out += ' '
            s_out += tok; last_was_num = True
    return s_out

new_word_d = transform_path(flat_word_d, slab_cx, slab_cy, scale, flat_cx, flat_cy)

# Replace the second (wordmark) path in the tilted SVG with the flat one.
matches = list(path_re.finditer(tilted))
word_match = matches[1]
new_word_tag = f'<path d="{new_word_d}" fill="{flat_word_fill}"'
result = tilted[:word_match.start()] + new_word_tag + tilted[word_match.end():]

with open(out_path, 'w') as f:
    f.write(result)
PY

echo "wrote $OUT"
