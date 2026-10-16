# 🗼 Lighthouse TUI

Unofficial lighthouse TUI app.

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/lighthouse-tui.svg)](https://pkg.go.dev/github.com/yourusername/lighthouse-tui)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/lighthouse-tui)](https://goreportcard.com/report/github.com/yourusername/lighthouse-tui)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A blazing-fast, terminal-native interactive dashboard for auditing website performance, accessibility, SEO, and best practices. Built in **Go** using Charm's [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

---

## ✨ Key Features

- ⚡ **Terminal-Native Speed:** Asynchronous execution with real-time UI updates powered by Go concurrency.
- 🎨 **Beautiful UI Layouts:** Styled using [Lip Gloss](https://github.com/charmbracelet/lipgloss) with custom color palettes calibrated for dark and light terminals.
- ⌨️ **Vim-Inspired Navigation:** Native keyboard controls (`j`/`k`, `/` search, tab navigation) and full mouse wheel scrolling.
- 📊 **Detailed Breakdowns:** Deep-dive into specific score metrics (LCP, CLS, FCP, TBT) with audit explanations and remediation tips.
- 📱 **Multi-Device Profiles:** Quickly toggle between Mobile and Desktop emulation profiles.
- 💾 **Export Capabilities:** Save audit summaries directly as JSON or Markdown reports.

---

## 📋 Prerequisites

Before running **Lighthouse TUI**, ensure you have:

- [Go](https://go.dev/) (1.21 or later)
- [Node.js](https://nodejs.org/) with `lighthouse` installed globally:
  ```bash
  npm install -g lighthouse
  ```
- Google Chrome or Chromium installed locally (required by Lighthouse CLI).

---

## 📦 Installation

### Via `go install`

```bash
go install github.com/wojtek0123/lighthouse-tui@latest
```

### From Source

```bash
# Clone the repository
git clone https://github.com/wojtek0123/lighthouse-tui.git
cd lighthouse-tui

# Build the binary
go build -o lighthouse-tui cmd/lighthouse-tui/main.go

# (Optional) Move binary to your system PATH
mv lighthouse-tui /usr/local/bin/
```

---

## 🛠️ Architecture & Tech Stack

This TUI strictly follows **The Elm Architecture** using the [Charm](https://charm.sh/) ecosystem:

- **Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **UI Components:** [Bubbles](https://github.com/charmbracelet/bubbles)

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for details.
