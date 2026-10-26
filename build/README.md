# Build Directory

Build configuration and packaging assets. Nothing here is generated at build time — every
file is kept by hand, and the release workflow stamps the version into the ones that carry
it.

* `bin/` — build output (gitignored)
* `config.yml` — `dev_mode` for `wails3 dev`: regenerate bindings, build, run against Vite
* `appicon.png`, `appicon.svg` — the source icon

## Windows

- `icon.ico` — the executable and installer icon.
- `wails.exe.manifest`, `info.json` — compiled into the executable as a `.syso` by
  `wails3 generate syso`. The release writes the tag's version into `info.json`.
- `installer/project.nsi` — the NSIS installer, including the uninstaller's opt-in
  *Campaigns, encounters and homebrew* tick-box. `installer/wails_tools.nsh` holds the
  macros it uses; the version comes in as `-DINFO_PRODUCTVERSION`. The WebView2 bootstrapper
  is written to `installer/tmp` by `wails3 generate webview2bootstrapper`.

## macOS

- `Info.plist` — copied into `DH-Companion.app/Contents`. The release writes the tag's
  version into it; the `.icns` is made from `appicon.png` with `sips` and `iconutil`.

## Linux

- `package.sh` — builds the `.deb` and the AppImage from `bin/DH-Companion`.
- `hilt.desktop` — the menu entry both packages install.
