# BetterManPage (bman)

A lightweight, modern CLI utility in Go to enhance the reading experience of Unix `man` pages. It acts as a wrapper around the system `man` command, providing syntax highlighting, easy navigation, and search capabilities within a terminal user interface (TUI).

## Features

-   **Syntax Highlighting:** Automatically colorizes headers, command options (flags), placeholders (`<value>`, `[FILE]`), and variable names.
-   **Section Navigation (TOC):** A toggleable sidebar to quickly jump between standard sections like `NAME`, `SYNOPSIS`, `DESCRIPTION`, and `EXAMPLES`.
-   **Vim-style Search:** Search within the page using `/`, and navigate matches with `n` and `N`.
-   **Responsive TUI:** Clean interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Installation

### Prerequisites

-   Go 1.21 or later
-   Standard `man` and `col` utilities (available on most Unix-like systems)

### Build from Source

```bash
git clone https://github.com/yourusername/bettermanpage.git
cd bettermanpage
go build -o bman .
sudo mv bman /usr/local/bin/
```

## Usage

Run `bman` followed by the name of the man page you want to view:

```bash
bman ls
bman tar
bman git-commit
```

### Keybindings

-   `Tab`: Toggle Table of Contents (TOC)
-   `Up/Down` or `k/j`: Scroll content or navigate TOC
-   `Enter`: Jump to the selected section in TOC
-   `/`: Start search
-   `n`: Next search match
-   `N`: Previous search match
-   `q` or `Esc`: Quit

## Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o bman .
```

## License

MIT
