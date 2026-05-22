# easesee

A terminal dashboard for managing locally registered dev servers. Backed by a single Go binary, distributed via npm for one-line install.

> See the [main project README](https://github.com/hayoung123/easesee) for screenshots, full key bindings, and design notes.

## Install

```bash
npm install -g easesee
# or
pnpm add -g easesee
```

On first run, the package downloads the appropriate native binary from the matching [GitHub release](https://github.com/hayoung123/easesee/releases) for your platform.

## Run

```bash
easesee            # launch the TUI
easesee register --help
easesee ls
```

## Supported platforms

- macOS (Intel + Apple Silicon)
- Linux (x86_64 + arm64)

## License

MIT
