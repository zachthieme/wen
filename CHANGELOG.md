# Changelog

### v0.3.0 — March 17, 2026: Calendar Enhancements

**Week numbers:**
- `w` toggles week number column on/off
- US and ISO 8601 numbering systems (configurable)
- `week_start_day` controls grid layout (Sunday or Monday)

**Color themes:**
- Four built-in themes: default, catppuccin-mocha, dracula, nord
- Custom color overrides via config

**New keybindings:**
- Arrow keys mirror vim keys for day/week navigation
- `J`/`K` for year jump (with leap day clamping)
- `t` to jump to today
- `y` to yank cursor date to clipboard (pbcopy, wl-copy, xclip, xsel)
- `?` to toggle help bar

**Config file:**
- `~/.config/wen/config.yaml` (auto-created on first run)
- Respects `$XDG_CONFIG_HOME`
- Invalid config falls back to defaults with a warning

### v0.2.0 — March 17, 2026: Interactive Calendar & Rename

**Renamed from `zdate` to `wen`.**

**Interactive calendar:**
- `wen cal` launches an interactive terminal calendar at the current month
- `wen cal december 2026` starts the calendar at a specific month
- Vim-style navigation: `h`/`l` (day), `j`/`k` (week), `H`/`L` (month)
- `Enter` prints selected date to stdout, `q`/`Esc` cancels (exit 1)
- Today highlighted with bold/underline, cursor with reverse video
- Day/week navigation wraps across month boundaries
- Month jumps clamp day number (e.g., Jan 31 → Feb 28)
- Built with Bubbletea and Lipgloss

**Breaking changes:**
- `--now` flag removed
- Binary renamed from `zdate` to `wen`
- GitHub repo moved to `zachthieme/wen`

### v0.1.0 — March 17, 2026: Initial Release

- Natural language date parsing via `olebedev/when`
- Three input modes: positional args, piped stdin, no-args (prints today)
- Output always `yyyy-mm-dd`
- Error messages to stderr with exit code 1
- Nix flake with binary and source builds
- GoReleaser for cross-platform releases (linux/darwin x amd64/arm64)
- CI pipeline with test and lint

---

For usage details, see [README](README.md).
