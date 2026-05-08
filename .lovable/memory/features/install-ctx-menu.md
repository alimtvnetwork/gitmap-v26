---
name: Install Ctx Menu
description: Cross-platform right-click context menu installer (Windows registry, macOS Quick Actions, Linux Nautilus/Dolphin/Thunar)
type: feature
---

`gitmap install ctx` / `gitmap uninstall ctx` adds gitmap actions to the OS right-click menu.

**Single source of truth**: `ctxMenu()` in `gitmap/cmd/installctxentries.go` returns a nested `[]ctxEntry` tree (Scan / Clone / Release / Repos / Visibility / Tools + top-level Open-terminal + Docs). All platforms read from this same table.

**Windows** (`installctx.go` + `installctxmenu.go`): nested HKCU cascade under `Software\Classes\Directory\{Background,}\shell\gitmap` using `MUIVerb` + empty `SubCommands` pattern. Real submenus.

**macOS** (`installctxmac.go`): one `.workflow` bundle per flat entry under `~/Library/Services`. Minimal `Info.plist` + `document.wflow` (XML plist) wrapping a `Run Shell Script` action. Shows in Finder Quick Actions / Services. After install, `pkill -KILL -u $USER cfprefsd` refreshes Finder.

**Linux** (`installctxlinux.go` + `installctxlinuxthunar.go`):
- Nautilus: shell script per entry in `~/.local/share/nautilus/scripts/gitmap/`.
- Dolphin: single `.desktop` with `X-KDE-Submenu=gitmap` + `Actions=` listing every entry → real cascade.
- Thunar: marker-delimited (`<!-- gitmap-ctx-begin/end -->`) `<action>` block in `~/.config/Thunar/uca.xml`; uninstall strips block in place.

**Flatten helper** (`installctxflatten.go`): on macOS/Linux, `Category ▸ Child` collapses to `gitmap: Category — Child`. Slug = lowercase a-z0-9 with `-` separators (filesystem-safe id for workflow folder / .desktop action / nautilus filename).

**Three execution modes** (per `constants.CtxMode`):
- `CtxModeTerminal` — opens terminal at folder, runs `gitmap <args>`, keeps window open. Default for mutating commands.
- `CtxModeSilent` — runs hidden, surfaces output via `msg.exe` (Win) / `display notification` (mac) / `notify-send` (Linux). Used for read-only queries (`list-versions`, `find-next`, `*-repos`, `docs`).
- `CtxModePrefill` — opens terminal at folder with `gitmap ` prompt waiting (Open-terminal-here entry only).

**Constraints**:
- Each impl file ≤200 lines, functions ≤15 lines, no string literals (everything in `constants_installctx.go` + `constants_installctx_unix.go`).
- HKCU only on Windows (no admin); `$HOME` only on Unix (no sudo).
- Excluded from `install all` — opt-in only (changes shell chrome).
- Uninstall is scoped: never wildcards parent keys, only strips marker-delimited block in Thunar.

Spec: `spec/04-generic-cli/30-install-ctx.md`.
