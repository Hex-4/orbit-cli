# Orbit CLI 🚀

A sleek, highly extensible CLI status dashboard and automation toolkit built in Go.

## Overview

Orbit is designed for developers who love the terminal and want a central hub for their automation tasks and system telemetry. Inspired by the creator of 'bramble', Orbit aims to be lightweight, modular, and AI-ready.

## Features

- **TUI Dashboard**: Built with `Bubble Tea` for a beautiful terminal experience.
- **Bramble Integration**: Hook into ambient agent frameworks seamlessly.
- **Modular Design**: Easy to extend with Go-based automation scripts.
- **Telemetry**: Real-time system status monitoring.

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
git clone https://github.com/Hex-4/orbit-cli.git
cd orbit-cli
go mod tidy
```

### Running the Dashboard

```bash
go run main.go
```

### Running Automation Tasks

```bash
go run main.go run
```

## Future Roadmap

- [ ] Real-time integration with Poke's ingest endpoints.
- [ ] Dynamic plugin loading via `.so` files.
- [ ] Advanced TUI widgets for network and disk I/O.

---

*Built with ❤️ for Hex-4.*
