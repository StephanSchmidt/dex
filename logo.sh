#!/usr/bin/env bash
# Regenerate the dex wordmark. Calls the sibling ../3dlogo tool with the
# exact parameters that produced the current logo, then rasterises to a
# 720×360 PNG via Inkscape.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
THREEDLOGO_DIR="${REPO_ROOT}/../3dlogo"
FONT="${HOME}/Fonts/BerkeleyMono-SemiBold-Condensed.otf"
GRADIENT='linear-gradient(180deg, #2d1004 0%, #c1272d 50%, #ff8c5a 100%)'
SVG="${REPO_ROOT}/assets/dex-logo.svg"
PNG="${REPO_ROOT}/assets/dex-logo.png"

[[ -d "${THREEDLOGO_DIR}" ]] || { echo "3dlogo not found at ${THREEDLOGO_DIR}" >&2; exit 1; }
[[ -f "${FONT}" ]]            || { echo "font not found at ${FONT}" >&2; exit 1; }

# Build the 3dlogo binary if missing.
if [[ ! -x "${THREEDLOGO_DIR}/3dlogo" ]]; then
  (cd "${THREEDLOGO_DIR}" && go build -o 3dlogo)
fi

mkdir -p "$(dirname "${SVG}")"

(cd "${THREEDLOGO_DIR}" && ./3dlogo \
  -font "${FONT}" \
  -text dex \
  -outline 8 \
  -flat-bottom=true \
  -gradient "${GRADIENT}" \
  -out "${SVG}")

# 720×360 PNG matches the loupe raster dimensions for consistency.
inkscape "${SVG}" -w 720 -h 360 -o "${PNG}"

echo "wrote:"
ls -la "${SVG}" "${PNG}"
