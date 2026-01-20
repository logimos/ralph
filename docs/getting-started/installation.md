# Installation

Ralph can be installed via Go CLI or built from source.

## Prerequisites

- Go 1.21 or later
- An AI agent CLI tool (Cursor Agent, Claude, etc.)
- Git (for commit functionality)

## Install via Go CLI (Recommended)

Install the latest version directly from GitHub:

```bash
go install github.com/start-it/ralph@latest
```

Install a specific version:

```bash
go install github.com/start-it/ralph@v1.0.0
```

Install from the main branch:

```bash
go install github.com/start-it/ralph@main
```

After installation, make sure `$GOPATH/bin` or `$GOBIN` is in your PATH. The binary will be available as `ralph`.

## Verify Installation

```bash
ralph -version
# Output: ralph version v1.x.x
```

## Build from Source

Clone the repository and build:

```bash
git clone https://github.com/start-it/ralph.git
cd ralph
make build
```

The binary will be created in the current directory.

## Local Development Install

For development work:

```bash
make install-local
```

This builds and installs the current version to `$GOPATH/bin` or `$GOBIN` for testing local changes.

## Download Pre-built Binaries

Pre-built binaries are available on the [GitHub Releases](https://github.com/start-it/ralph/releases) page for:

- Linux (amd64, arm64)
- macOS (Intel & Apple Silicon)
- Windows (amd64)

```bash
# Example: Linux amd64
curl -L https://github.com/start-it/ralph/releases/latest/download/ralph-linux-amd64 -o ralph
chmod +x ralph
sudo mv ralph /usr/local/bin/
```

## Next Steps

- [Quick Start Guide](quickstart.md) - Get running in 5 minutes
- [Configuration](configuration.md) - Customize Ralph for your project
