# BetterManPage Implementation Plan (Phase 2 & Beyond)

This document outlines the roadmap and technical specifications for enhancing the `bettermanpage` utility based on proposed ideas.

## 1. Roadmap

### Phase 5: Enhanced Navigation & Vim Parity
- **Vim Keybindings:** Support `gg`, `G`, `d`, `u`, `ctrl+f`, `ctrl+b`.
- **Search Enhancements:** Show match count (e.g., `[1/5]`) and highlight current match in the viewport.
- **Section Breadcrumbs:** Display current section name at the top of the viewport during scroll.

### Phase 6: Customization & Configuration
- **Config File:** Implement YAML-based configuration (`~/.config/bman/config.yaml`).
- **Theming:** Support custom color schemes (e.g., Catppuccin, Nord, Solarized).
- **Aliases:** Add a command (`bman --alias`) to generate shell alias scripts for `.zshrc`/`.bashrc`.

### Phase 7: External Integrations (Partially Implemented)
- **TLDR Overlay [DONE]:** Press `e` to fetch and display concise examples from [tldr.sh](https://tldr.sh/).
- **Cross-Page Links:** Make "SEE ALSO" references and command names selectable to jump to their respective man pages.

### Phase 8: Performance & Ecosystem
- **Caching:** Cache parsed and highlighted content in `~/.cache/bman/` for faster subsequent loads.
- **Robust Parsing:** Transition from regex-based highlighting to a more structured approach using `mandoc -T utf8` or `groff`.

---

## 2. Technical Specifications

### 2.1 Configuration (`config.yaml`)
```yaml
theme:
  header: "#F5C2E7"
  option: "#A6E3A1"
  value: "#FAB387"
  toc_selected: "#F38BA8"

keybindings:
  toggle_toc: "tab"
  search: "/"
  tldr: "e"

features:
  caching: true
  history_limit: 50
```

### 2.2 Caching Strategy
- **Key:** Command name + Section (e.g., `ls.1`).
- **Store:** Raw ANSI-highlighted string and pre-parsed section list.
- **Invalidation:** Check file `mtime` of the system man page file if possible, or simple TTL.

### 2.3 TLDR Integration
- **Mechanism:** Use `http` package to fetch `https://tldr.sh/assets/pages/common/ls.md` or use a dedicated Go library.
- **UI:** Display in a centered popup modal (using `lipgloss` layering).

### 2.4 Vim Navigation Logic
| Key | Action |
|-----|--------|
| `gg` | `m.viewport.GotoTop()` |
| `G` | `m.viewport.GotoBottom()` |
| `d` | `m.viewport.HalfPageDown()` |
| `u` | `m.viewport.HalfPageUp()` |
| `ctrl+f` | `m.viewport.PageDown()` |
| `ctrl+b` | `m.viewport.PageUp()` |

---

## 3. Implementation Priorities (Immediate Next Steps)

1.  **Vim Keybindings:** High impact, low effort. Improves user muscle memory immediately.
2.  **Match Count in Search:** Critical for search usability.
3.  **TLDR Overlay:** Adds significant value beyond standard `man`.
4.  **Config File Support:** Necessary for theming and personal preference.

---

## 4. Verification Plan
- **Vim Keys:** Test all navigation keys in large pages (e.g., `git`).
- **Search:** Verify `n/N` correctly cycles through all matches and the counter updates.
- **TLDR:** Mock API response and ensure the modal displays correctly without crashing the TUI.
