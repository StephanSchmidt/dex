#!/usr/bin/env bash
# Regenerate assets/dex-logo.svg via ../3dlogo.
#
# Requires:
#   - ../3dlogo built (see github.com/StephanSchmidt/3dlogo)
#   - BricolageGrotesque96ptCondensed-ExtraBold.ttf in ~/Fonts/
#
# Style matches loupe's vertical 3-stop gradient — dark → saturated → light
# (top to bottom). Default pitch/yaw/size from 3dlogo are kept.

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
LOGO_BIN="${LOGO_BIN:-$REPO_DIR/../3dlogo/3dlogo}"
FONT="${FONT:-$HOME/Fonts/BricolageGrotesque96ptCondensed-ExtraBold.ttf}"
OUT="${OUT:-$REPO_DIR/assets/dex-logo.svg}"

mkdir -p "$(dirname "$OUT")"

"$LOGO_BIN" \
  -font "$FONT" \
  -text "dex" \
  -gradient "linear-gradient(180deg, #2d1004 0%, #c1272d 50%, #ff8c5a 100%)" \
  -fg "#ffffff" \
  -out "$OUT"
