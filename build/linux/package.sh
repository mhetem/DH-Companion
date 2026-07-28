#!/usr/bin/env bash
# Builds a .deb and an AppImage from an already-compiled build/bin/DH-Companion.
#
# Usage: build/linux/package.sh <version> [outdir]
#   version   e.g. v1.2.0 or 1.2.0 — a leading v is stripped for package metadata
#   outdir    where the artifacts land (default: dist)
#
# Run `wails build -tags webkit2_41` first. Needs dpkg-deb for the .deb; the
# AppImage step downloads appimagetool and is skipped if that download fails, so a
# machine without network still gets the .deb.
set -euo pipefail

VERSION_RAW="${1:?usage: package.sh <version> [outdir]}"
VERSION="${VERSION_RAW#v}"
OUTDIR="${2:-dist}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="$ROOT/build/bin/DH-Companion"
DESKTOP="$ROOT/build/linux/hilt.desktop"
ICON_SVG="$ROOT/build/appicon.svg"
ICON_PNG="$ROOT/build/appicon.png"

[ -x "$BIN" ] || { echo "missing $BIN — run wails build first" >&2; exit 1; }

mkdir -p "$ROOT/$OUTDIR"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# ── shared payload ───────────────────────────────────────────────────────────
# The SVG goes to hicolor/scalable, which is authoritative and size-free. The PNG
# goes to pixmaps rather than a hicolor size directory: appicon.png is 1024x1024,
# and dropping it in, say, 512x512/apps would claim a size it isn't. pixmaps is the
# legacy fallback and carries no size contract.
stage() {
  local root="$1" bindir="$2"
  install -Dm755 "$BIN"       "$root$bindir/hilt"
  install -Dm644 "$DESKTOP"   "$root/usr/share/applications/hilt.desktop"
  install -Dm644 "$ICON_SVG"  "$root/usr/share/icons/hicolor/scalable/apps/hilt.svg"
  install -Dm644 "$ICON_PNG"  "$root/usr/share/pixmaps/hilt.png"
  install -Dm644 "$ROOT/LICENSE"   "$root/usr/share/doc/hilt/LICENSE"
  install -Dm644 "$ROOT/NOTICE.md" "$root/usr/share/doc/hilt/NOTICE.md"
}

# ── .deb ─────────────────────────────────────────────────────────────────────
DEB="$WORK/deb"
stage "$DEB" /usr/bin
mkdir -p "$DEB/DEBIAN"
cat > "$DEB/DEBIAN/control" <<EOF
Package: hilt
Version: $VERSION
Section: games
Priority: optional
Architecture: amd64
Depends: libgtk-3-0, libwebkit2gtk-4.1-0
Maintainer: mhetem <284336022+mhetem@users.noreply.github.com>
Homepage: https://github.com/mhetem/DH-Companion
Description: Daggerheart Compatible desktop companion
 Build encounters against a live battle-point budget, run the fight with
 per-adversary HP, Stress and Fear tracking, and keep the campaign between
 sessions with a session log, typed notes and full-text search.
 .
 Local-first: everything lives in one SQLite file on your machine.
EOF

dpkg-deb --build --root-owner-group "$DEB" "$ROOT/$OUTDIR/Hilt-${VERSION}-amd64.deb" >/dev/null
echo "built $OUTDIR/Hilt-${VERSION}-amd64.deb"

# ── AppImage ─────────────────────────────────────────────────────────────────
APPDIR="$WORK/Hilt.AppDir"
stage "$APPDIR" /usr/bin
# appimagetool wants the desktop file and icon at the AppDir root as well.
cp "$DESKTOP"  "$APPDIR/hilt.desktop"
cp "$ICON_SVG" "$APPDIR/hilt.svg"
cp "$ICON_PNG" "$APPDIR/hilt.png"
cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/hilt" "$@"
EOF
chmod +x "$APPDIR/AppRun"

TOOL="$WORK/appimagetool"
if curl -fsSL -o "$TOOL" \
  "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage"; then
  chmod +x "$TOOL"
  # --appimage-extract-and-run avoids needing FUSE, which CI runners lack.
  if ARCH=x86_64 "$TOOL" --appimage-extract-and-run "$APPDIR" \
      "$ROOT/$OUTDIR/Hilt-${VERSION}-x86_64.AppImage" >/dev/null 2>&1; then
    echo "built $OUTDIR/Hilt-${VERSION}-x86_64.AppImage"
  else
    echo "appimagetool failed — skipping the AppImage" >&2
  fi
else
  echo "could not download appimagetool — skipping the AppImage" >&2
fi
